/*
Copyright © 2021 Pete Cornish <outofcoffee@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package docker

import (
	"context"
	"fmt"
	"gatehill.io/imposter/internal/debounce"
	"gatehill.io/imposter/internal/engine"
	"gatehill.io/imposter/internal/logging"
	"gatehill.io/imposter/internal/plugin"
	"gatehill.io/imposter/internal/stringutil"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const containerConfigDir = "/opt/imposter/config"
const containerPluginDir = "/opt/imposter/plugins"
const containerFileCacheDir = "/tmp/imposter-cache"
const removalTimeoutSec = 5

var logger = logging.GetLogger()

func (d *DockerMockEngine) Start(wg *sync.WaitGroup) bool {
	return d.startWithOptions(wg, d.options)
}

func (d *DockerMockEngine) startWithOptions(wg *sync.WaitGroup, options engine.StartOptions) (success bool) {
	logger.Infof("starting mock engine on port %d - press ctrl+c to stop", options.Port)
	ctx, cli, err := buildCliClient()
	if err != nil {
		logger.Fatal(err)
	}

	if !d.provider.Satisfied() {
		if err := d.provider.Provide(engine.PullIfNotPresent); err != nil {
			logger.Fatal(err)
		}
	}

	mockHash, containerLabels := generateMetadata(d, options)

	if options.ReplaceRunning {
		stopDuplicateContainers(d, cli, ctx, mockHash)
	}

	// if not specified, falls back to default in container image
	containerUser := viper.GetString("docker.containerUser")
	logger.Tracef("container user: %s", containerUser)

	exposedPorts, portBindings := buildPorts(options)
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: d.provider.imageAndTag,
		Cmd: []string{
			"--configDir=" + containerConfigDir,
			fmt.Sprintf("--listenPort=%d", options.Port),
		},
		Env:          buildEnv(options),
		ExposedPorts: exposedPorts,
		Labels:       containerLabels,
		User:         containerUser,
	}, &container.HostConfig{
		Binds:        buildBinds(d, options),
		PortBindings: portBindings,
	}, nil, nil, "")
	if err != nil {
		logger.Fatal(err)
	}

	containerId := resp.ID
	d.debouncer.Register(wg, containerId)
	if err := cli.ContainerStart(ctx, containerId, container.StartOptions{}); err != nil {
		logger.Fatalf("error starting mock engine container: %v", err)
	}
	logger.Trace("starting Docker mock engine")

	d.containerId = containerId
	if err = streamLogsToStdIo(cli, ctx, containerId); err != nil {
		logger.Warn(err)
	}
	up := engine.WaitUntilUp(options.Port, d.shutDownC)

	// watch in case container stops
	go notifyOnStopBlocking(d, wg, containerId, cli, ctx)

	return up
}

func buildPorts(options engine.StartOptions) (nat.PortSet, nat.PortMap) {
	ports := map[int]int{
		options.Port: options.Port,
	}
	if options.DebugMode {
		ports[engine.DefaultDebugPort] = engine.DefaultDebugPort
	}

	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}
	for hp, cp := range ports {
		containerPort := nat.Port(fmt.Sprintf("%d/tcp", cp))
		hostPort := fmt.Sprintf("%d", hp)

		exposedPorts[containerPort] = struct{}{}
		portBindings[containerPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
	}
	return exposedPorts, portBindings
}

func buildEnv(options engine.StartOptions) []string {
	env := engine.BuildEnv(options, engine.EnvOptions{IncludeHome: false, IncludePath: false})
	if options.EnableFileCache {
		env = append(env, "IMPOSTER_CACHE_DIR=/tmp/imposter-cache", "IMPOSTER_OPENAPI_REMOTE_FILE_CACHE=true")
	}
	logger.Tracef("engine environment: %v", env)
	return env
}

func buildBinds(d *DockerMockEngine, options engine.StartOptions) []string {
	binds := []string{
		d.configDir + ":" + containerConfigDir + viper.GetString("docker.bindFlags"),
	}
	if options.EnablePlugins {
		logger.Tracef("plugins are enabled")
		pluginDir, err := plugin.EnsurePluginDir(options.Version)
		if err != nil {
			logger.Fatal(err)
		}
		binds = append(binds, pluginDir+":"+containerPluginDir)
	} else {
		logger.Tracef("plugins are disabled")
	}
	if options.EnableFileCache {
		logger.Tracef("file cache enabled")
		fileCacheDir, err := engine.EnsureFileCacheDir()
		if err != nil {
			logger.Fatal(err)
		}
		binds = append(binds, fileCacheDir+":"+containerFileCacheDir)
	} else {
		logger.Tracef("file cache disabled")
	}
	binds = append(binds, parseDirMounts(options.DirMounts)...)
	logger.Tracef("using binds: %v", binds)
	return binds
}

// parseDirMounts validates the directory mounts, generating
// the container path if not provided
func parseDirMounts(dirMounts []string) []string {
	var binds []string
	for _, mountSpec := range dirMounts {
		var hostDir string
		if strings.Contains(mountSpec, ":") {
			splitSpec := strings.Split(mountSpec, ":")
			hostDir = splitSpec[0]

		} else {
			hostDir = mountSpec
			// generate container path based on last dir name
			_, dir := filepath.Split(mountSpec)
			containerDir := filepath.Join("/opt/imposter/", dir)
			mountSpec = fmt.Sprintf("%s:%s", hostDir, containerDir)
		}

		hostDirInfo, err := os.Stat(hostDir)
		if err != nil {
			logger.Fatalf("failed to stat host dir: %s", hostDir)
		}
		if !hostDirInfo.IsDir() {
			logger.Fatalf("host path: %s is not a directory", hostDir)
		}
		binds = append(binds, mountSpec)
	}
	return binds
}

func generateMetadata(d *DockerMockEngine, options engine.StartOptions) (string, map[string]string) {
	absoluteConfigDir, _ := filepath.Abs(d.configDir)

	var mockHash string
	if options.Deduplicate != "" {
		mockHash = stringutil.Sha1hashString(options.Deduplicate)
	} else {
		mockHash = genDefaultHash(absoluteConfigDir, options.Port)
	}

	containerLabels := map[string]string{
		labelKeyManaged: "true",
		labelKeyDir:     absoluteConfigDir,
		labelKeyPort:    strconv.Itoa(options.Port),
		labelKeyHash:    mockHash,
	}
	return mockHash, containerLabels
}

func streamLogsToStdIo(cli *client.Client, ctx context.Context, containerId string) error {
	return streamLogs(cli, ctx, containerId, os.Stdout, os.Stderr)
}

func streamLogs(cli *client.Client, ctx context.Context, containerId string, outStream io.Writer, errStream io.Writer) error {
	containerLogs, err := cli.ContainerLogs(ctx, containerId, container.LogsOptions{
		ShowStdout: true,
		Follow:     true,
	})
	if err != nil {
		return fmt.Errorf("error streaming container logs for container with ID: %v: %v", containerId, err)
	}
	go func() {
		_, err := stdcopy.StdCopy(outStream, errStream, containerLogs)
		if err != nil {
			logger.Warnf("error streaming container logs for container with ID: %v: %v", containerId, err)
		}
	}()
	return nil
}

func buildCliClient() (context.Context, *client.Client, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, nil, err
	}
	return ctx, cli, nil
}

func (d *DockerMockEngine) StopImmediately(wg *sync.WaitGroup) {
	go func() { d.shutDownC <- true }()
	d.Stop(wg)
}

func (d *DockerMockEngine) Stop(wg *sync.WaitGroup) {
	if len(d.containerId) == 0 {
		logger.Tracef("no container ID to remove")
		wg.Done()
		return
	}
	if logger.IsLevelEnabled(logrus.TraceLevel) {
		logger.Tracef("stopping mock engine container %v", d.containerId)
	} else {
		logger.Info("stopping mock engine")
	}

	oldContainerId := d.containerId

	// supervisor to work-around removal race
	go func() {
		time.Sleep(removalTimeoutSec * time.Second)
		logger.Tracef("fired timeout supervisor for container %v removal", oldContainerId)
		d.debouncer.Notify(wg, debounce.AtMostOnceEvent{Id: oldContainerId})
	}()

	removeContainer(d, wg, oldContainerId)
}

func (d *DockerMockEngine) Restart(wg *sync.WaitGroup) {
	wg.Add(1)
	d.Stop(wg)

	// don't pull again
	restartOptions := d.options
	restartOptions.PullPolicy = engine.PullSkip

	d.startWithOptions(wg, restartOptions)
	wg.Done()
}

func (d *DockerMockEngine) ListAllManaged() ([]engine.ManagedMock, error) {
	cli, ctx, err := buildCliClient()
	if err != nil {
		logger.Fatal(err)
	}

	labels := map[string]string{
		labelKeyManaged: "true",
	}
	containers, err := findContainersWithLabels(ctx, cli, labels)
	if err != nil {
		logger.Fatalf("error searching for existing containers: %v", err)
	}
	return containers, nil
}

func (d *DockerMockEngine) StopAllManaged() int {
	cli, ctx, err := buildCliClient()
	if err != nil {
		logger.Fatal(err)
	}

	labels := map[string]string{
		labelKeyManaged: "true",
	}
	return stopContainersWithLabels(d, ctx, cli, labels)
}

func (d *DockerMockEngine) GetVersionString() (string, error) {
	if !d.provider.Satisfied() {
		if err := d.provider.Provide(engine.PullSkip); err != nil {
			return "", err
		}
	}

	output := new(strings.Builder)
	errOutput := new(strings.Builder)

	ctx, cli, err := buildCliClient()
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: d.provider.imageAndTag,
		Cmd: []string{
			"--version",
		},
	}, &container.HostConfig{}, nil, nil, "")
	if err != nil {
		return "", err
	}
	containerId := resp.ID

	wg := &sync.WaitGroup{}
	d.debouncer.Register(wg, containerId)
	if err := cli.ContainerStart(ctx, containerId, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("error starting mock engine container: %v", err)
	}
	if err = streamLogs(cli, ctx, containerId, output, errOutput); err != nil {
		return "", fmt.Errorf("error getting mock engine output: %v", err)
	}
	notifyOnStopBlocking(d, wg, containerId, cli, ctx)
	return engine.SanitiseVersionOutput(output.String()), nil
}
