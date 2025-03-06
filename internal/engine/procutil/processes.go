package procutil

import (
	"fmt"
	"gatehill.io/imposter/internal/logging"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gatehill.io/imposter/internal/engine"
	"github.com/shirou/gopsutil/v4/process"
)

var logger = logging.GetLogger()

// ProcessMatcher defines how to identify a specific type of imposter process
type ProcessMatcher struct {
	// ProcessName is the name of the process to match (e.g., "java" or "imposter")
	ProcessName string
	// CommandPattern is a regex pattern to match against command line arguments
	CommandPattern string
	// GetPort is a function that determines the port from command line arguments
	GetPort func(cmdline []string, env []string) int
}

// FindImposterProcesses finds all imposter processes matching the given matcher
func FindImposterProcesses(matcher ProcessMatcher) ([]engine.ManagedMock, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("error listing processes: %v", err)
	}

	var mocks []engine.ManagedMock
	for _, p := range processes {
		cmdline, err := p.CmdlineSlice()
		if err != nil {
			continue
		}
		procName, err := p.Name()
		if err != nil {
			continue
		}
		if !isImposterProc(cmdline, procName, matcher) {
			continue
		}

		env, err := p.Environ()
		if err != nil {
			logger.Warnf("error reading environment variables for process %d: %v", p.Pid, err)
			env = []string{}
		}
		logger.Tracef("found %s Imposter process %d: %v, env: %v", matcher.ProcessName, p.Pid, cmdline, env)

		port := matcher.GetPort(cmdline, env)

		mock := engine.ManagedMock{
			ID:   fmt.Sprintf("%d", p.Pid),
			Name: procName,
			Port: port,
		}
		mocks = append(mocks, mock)
	}
	return mocks, nil
}

func isImposterProc(cmdline []string, procName string, matcher ProcessMatcher) bool {
	if procName != matcher.ProcessName {
		return false
	}
	for _, arg := range cmdline {
		if matched, _ := regexp.MatchString(matcher.CommandPattern, arg); matched {
			return true
		}
	}
	return false
}

// StopManagedProcesses stops all processes matching the given matcher
func StopManagedProcesses(matcher ProcessMatcher) (int, error) {
	processes, err := FindImposterProcesses(matcher)
	if err != nil {
		return 0, err
	}
	if len(processes) == 0 {
		return 0, nil
	}
	for _, proc := range processes {
		pid, err := strconv.Atoi(proc.ID)
		if err != nil {
			return 0, fmt.Errorf("invalid process ID: %v", err)
		}
		logger.Tracef("finding %s process to kill with PID: %d", matcher.ProcessName, pid)
		p, err := os.FindProcess(pid)
		if err != nil {
			return 0, fmt.Errorf("failed to find process: %v", err)
		}
		logger.Debugf("killing %s process with PID: %d", matcher.ProcessName, pid)
		err = p.Kill()
		if err != nil {
			logger.Warnf("error killing %s process with PID: %d: %v", matcher.ProcessName, pid, err)
		}
	}
	return len(processes), nil
}

// ReadArg parses the command line arguments to find the value of a given argument
func ReadArg(cmdline []string, longArg string, shortArg string) string {
	for i := range cmdline {
		arg := cmdline[i]
		if matched, _ := regexp.MatchString("--"+longArg+"=", arg); matched {
			// combined form: "--longArg=VAL"
			return strings.TrimPrefix(arg, "--"+longArg+"=")
		} else if (arg == "--"+longArg || arg == "-"+shortArg) && i < len(cmdline)-1 {
			// separate form: "--longArg", "VAL"
			return cmdline[i+1]
		}
	}
	return ""
}
