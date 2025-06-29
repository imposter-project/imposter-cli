package golang

import (
	"fmt"
	"gatehill.io/imposter/internal/debounce"
	"gatehill.io/imposter/internal/engine"
	"gatehill.io/imposter/internal/engine/procutil"
	"gatehill.io/imposter/internal/logging"
	"gatehill.io/imposter/internal/plugin"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

var logger = logging.GetLogger()

// GolangMockEngine implements the MockEngine interface for the golang implementation
type GolangMockEngine struct {
	configDir string
	options   engine.StartOptions
	provider  *Provider
	cmd       *exec.Cmd
	debouncer debounce.Debouncer
	shutDownC chan bool
}

// NewGolangMockEngine creates a new instance of the golang mock engine
func NewGolangMockEngine(configDir string, options engine.StartOptions, provider *Provider) *GolangMockEngine {
	return &GolangMockEngine{
		configDir: configDir,
		options:   options,
		provider:  provider,
		debouncer: debounce.Build(),
		shutDownC: make(chan bool),
	}
}

func (g *GolangMockEngine) Start(wg *sync.WaitGroup) bool {
	return g.startWithOptions(wg, g.options)
}

func (g *GolangMockEngine) startWithOptions(wg *sync.WaitGroup, options engine.StartOptions) (success bool) {
	if len(options.DirMounts) > 0 {
		logger.Warnf("golang engine does not support directory mounts - these will be ignored")
	}
	env := g.buildEnv(options)
	command := (*g.provider).GetStartCommand([]string{}, env)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Start(); err != nil {
		logger.Errorf("failed to start golang mock engine: %v", err)
		return false
	}
	g.debouncer.Register(wg, strconv.Itoa(command.Process.Pid))
	logger.Trace("starting golang mock engine")
	g.cmd = command

	// watch in case process stops
	up := engine.WaitUntilUp(options.Port, g.shutDownC)

	go g.notifyOnStopBlocking(wg)
	return up
}

func (g *GolangMockEngine) buildEnv(options engine.StartOptions) []string {
	env := engine.BuildEnv(options, engine.EnvOptions{IncludeHome: true, IncludePath: true})
	env = append(env,
		fmt.Sprintf("IMPOSTER_PORT=%d", options.Port),
		fmt.Sprintf("IMPOSTER_CONFIG_DIR=%s", g.configDir),
	)
	if options.EnablePlugins {
		logger.Tracef("plugins are enabled")
		pluginDir, err := plugin.EnsurePluginDir(options.Version)
		if err != nil {
			logger.Fatal(err)
		}
		env = append(env,
			"IMPOSTER_PLUGIN_DIR="+pluginDir,
			"IMPOSTER_EXTERNAL_PLUGINS=true",
		)
	} else {
		logger.Tracef("plugins are disabled")
	}
	if options.EnableFileCache {
		logger.Tracef("file cache not supported by golang engine")
	}
	logger.Tracef("engine environment: %v", env)
	return env

}

func (g *GolangMockEngine) Stop(wg *sync.WaitGroup) {
	if g.cmd == nil {
		logger.Tracef("no process to remove")
		wg.Done()
		return
	}
	if logger.IsLevelEnabled(logrus.TraceLevel) {
		logger.Tracef("stopping mock engine with PID: %v", g.cmd.Process.Pid)
	} else {
		logger.Info("stopping mock engine")
	}

	err := g.cmd.Process.Kill()
	if err != nil {
		logger.Fatalf("error stopping engine with PID: %d: %v", g.cmd.Process.Pid, err)
	}
	g.notifyOnStopBlocking(wg)
}

func (g *GolangMockEngine) StopImmediately(wg *sync.WaitGroup) {
	go func() { g.shutDownC <- true }()
	g.Stop(wg)
}

func (g *GolangMockEngine) Restart(wg *sync.WaitGroup) {
	wg.Add(1)
	g.Stop(wg)

	// don't pull again
	restartOptions := g.options
	restartOptions.PullPolicy = engine.PullSkip

	g.startWithOptions(wg, restartOptions)
	wg.Done()
}

func (g *GolangMockEngine) notifyOnStopBlocking(wg *sync.WaitGroup) {
	if g.cmd == nil || g.cmd.Process == nil {
		logger.Trace("no subprocess - notifying immediately")
		g.debouncer.Notify(wg, debounce.AtMostOnceEvent{})
	}
	pid := strconv.Itoa(g.cmd.Process.Pid)
	if g.cmd.ProcessState != nil && g.cmd.ProcessState.Exited() {
		logger.Tracef("process with PID: %v already exited - notifying immediately", pid)
		g.debouncer.Notify(wg, debounce.AtMostOnceEvent{Id: pid})
	}
	_, err := g.cmd.Process.Wait()
	if err != nil {
		g.debouncer.Notify(wg, debounce.AtMostOnceEvent{
			Id:  pid,
			Err: fmt.Errorf("failed to wait for process with PID: %v: %v", pid, err),
		})
	} else {
		g.debouncer.Notify(wg, debounce.AtMostOnceEvent{Id: pid})
	}
}

func (g *GolangMockEngine) ListAllManaged() ([]engine.ManagedMock, error) {
	return procutil.FindImposterProcesses(matcher)
}

func (g *GolangMockEngine) StopAllManaged() int {
	count, err := procutil.StopManagedProcesses(matcher)
	if err != nil {
		logger.Fatal(err)
	}
	return count
}

func (g *GolangMockEngine) GetVersionString() (string, error) {
	// TODO get from binary
	return g.options.Version, nil
}
