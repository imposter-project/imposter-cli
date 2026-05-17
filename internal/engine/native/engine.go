package native

import (
	"fmt"
	"github.com/imposter-project/imposter-cli/internal/debounce"
	"github.com/imposter-project/imposter-cli/internal/engine"
	"github.com/imposter-project/imposter-cli/internal/engine/procutil"
	"github.com/imposter-project/imposter-cli/internal/logging"
	"github.com/imposter-project/imposter-cli/internal/plugin"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

var logger = logging.GetLogger()

// NativeMockEngine implements the MockEngine interface for the native binary implementation
type NativeMockEngine struct {
	configDir string
	options   engine.StartOptions
	provider  *Provider
	cmd       *exec.Cmd
	debouncer debounce.Debouncer
	shutDownC chan bool
}

// NewNativeMockEngine creates a new instance of the native mock engine
func NewNativeMockEngine(configDir string, options engine.StartOptions, provider *Provider) *NativeMockEngine {
	return &NativeMockEngine{
		configDir: configDir,
		options:   options,
		provider:  provider,
		debouncer: debounce.Build(),
		shutDownC: make(chan bool),
	}
}

func (g *NativeMockEngine) Start(wg *sync.WaitGroup) bool {
	return g.startWithOptions(wg, g.options)
}

func (g *NativeMockEngine) startWithOptions(wg *sync.WaitGroup, options engine.StartOptions) (success bool) {
	if len(options.DirMounts) > 0 {
		logger.Warnf("native engine does not support directory mounts - these will be ignored")
	}
	env := g.buildEnv(options)
	command := (*g.provider).GetStartCommand([]string{}, env)
	if options.IsDetached() {
		f, err := procutil.OpenDetachLog(options.DetachLog)
		if err != nil {
			logger.Fatal(err)
		}
		command.Stdout = f
		command.Stderr = f
		command.SysProcAttr = procutil.DetachSysProcAttr()
	} else {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
	}

	if err := command.Start(); err != nil {
		logger.Errorf("failed to start native mock engine: %v", err)
		return false
	}
	g.debouncer.Register(wg, strconv.Itoa(command.Process.Pid))
	logger.Trace("starting native mock engine")
	g.cmd = command

	switch options.Detach {
	case engine.DetachNow:
		// do not wait for health, do not reap - the OS reparents the child
		return true
	case engine.DetachHealthy:
		// wait for health but do not reap - the OS reparents the child
		return engine.WaitUntilUp(options.Port, g.shutDownC)
	default:
		// watch in case process stops
		up := engine.WaitUntilUp(options.Port, g.shutDownC)

		go g.notifyOnStopBlocking(wg)
		return up
	}
}

func (g *NativeMockEngine) GetID() string {
	if g.cmd == nil || g.cmd.Process == nil {
		return ""
	}
	return strconv.Itoa(g.cmd.Process.Pid)
}

func (g *NativeMockEngine) buildEnv(options engine.StartOptions) []string {
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
		logger.Tracef("file cache not supported by native engine")
	}
	logger.Tracef("engine environment: %v", env)
	return env

}

func (g *NativeMockEngine) Stop(wg *sync.WaitGroup) {
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

func (g *NativeMockEngine) StopImmediately(wg *sync.WaitGroup) {
	go func() { g.shutDownC <- true }()
	g.Stop(wg)
}

func (g *NativeMockEngine) Restart(wg *sync.WaitGroup) {
	wg.Add(1)
	g.Stop(wg)

	// don't pull again
	restartOptions := g.options
	restartOptions.PullPolicy = engine.PullSkip

	g.startWithOptions(wg, restartOptions)
	wg.Done()
}

func (g *NativeMockEngine) notifyOnStopBlocking(wg *sync.WaitGroup) {
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

func (g *NativeMockEngine) ListAllManaged() ([]engine.ManagedMock, error) {
	return procutil.FindImposterProcesses(matcher)
}

func (g *NativeMockEngine) StopAllManaged() int {
	count, err := procutil.StopManagedProcesses(matcher)
	if err != nil {
		logger.Fatal(err)
	}
	return count
}

func (g *NativeMockEngine) GetVersionString() (string, error) {
	// TODO get from binary
	return g.options.Version, nil
}
