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

package engine

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/imposter-project/imposter-cli/internal/config"
)

// DetachMode controls whether and how `up` backgrounds the mock.
type DetachMode int

const (
	// DetachNone runs the mock in the foreground (default behaviour).
	DetachNone DetachMode = iota
	// DetachNow starts the mock and returns immediately without waiting
	// for it to become healthy.
	DetachNow
	// DetachHealthy starts the mock, waits for the healthcheck to pass,
	// then returns control to the caller.
	DetachHealthy
)

type StartOptions struct {
	Port            int
	Version         string
	PullPolicy      PullPolicy
	LogLevel        string
	ReplaceRunning  bool
	Deduplicate     string
	EnablePlugins   bool
	EnableFileCache bool
	Environment     []string
	DirMounts       []string
	DebugMode       bool
	Detach          DetachMode
	// DetachLog is the resolved absolute path that a detached process
	// engine writes stdout/stderr to. Unused by the docker engine.
	DetachLog string
}

// IsDetached reports whether the mock should be run in the background.
func (o StartOptions) IsDetached() bool {
	return o.Detach != DetachNone
}

// DefaultDetachLogPath returns the default log file path for a detached
// process-engine mock listening on the given port.
func DefaultDetachLogPath(port int) (string, error) {
	globalDir, err := config.GetGlobalConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(globalDir, "logs", fmt.Sprintf("imposter-%d.log", port)), nil
}

type EnvOptions struct {
	IncludeHome bool
	IncludePath bool
}

type PullPolicy int

const (
	PullSkip         PullPolicy = iota
	PullAlways       PullPolicy = iota
	PullIfNotPresent PullPolicy = iota
)

type MockEngine interface {
	Start(wg *sync.WaitGroup) (success bool)
	Stop(wg *sync.WaitGroup)
	StopImmediately(wg *sync.WaitGroup)
	Restart(wg *sync.WaitGroup)
	ListAllManaged() ([]ManagedMock, error)
	StopAllManaged() int

	// StopManaged stops the single managed mock identified by id (the same
	// value reported in ManagedMock.ID, i.e. the short container ID for
	// docker or the PID for process engines). Returns (true, nil) if the
	// mock was found and stopped; (false, nil) if no managed mock with
	// that id exists in this engine; (false, err) if the engine could
	// not be queried.
	StopManaged(id string) (bool, error)

	GetVersionString() (string, error)

	// GetID returns an identifier for the running mock: the container ID
	// for the docker engine, or the process PID for process engines.
	// Returns an empty string if the mock has not been started.
	GetID() string
}

type EngineMetadata struct {
	EngineType EngineType
	Version    string
}

type Provider interface {
	Satisfied() bool
	Provide(policy PullPolicy) error
	GetEngineType() EngineType
	Build(configDir string, startOptions StartOptions) MockEngine

	// Bundle creates a single archive file containing the engine binary and
	// configuration files. The archive is written to the specified destination.
	// If the engine type is 'docker', destination should be a valid image name.
	Bundle(configDir string, dest string) error
}

type EngineLibrary interface {
	CheckPrereqs() (bool, []string)
	List() ([]EngineMetadata, error)
	GetProvider(version string) Provider

	// IsSealedDistro indicates whether a library represents a fixed distribution.
	// Fixed distributions have a single version, so do not support version
	// resolution or fetching engine binaries.
	IsSealedDistro() bool

	// ShouldEnsurePlugins indicates whether missing default plugins should be
	// installed before starting the engine.
	ShouldEnsurePlugins() bool
}

type MockHealth string

const (
	MockHealthHealthy   MockHealth = "healthy"
	MockHealthUnhealthy MockHealth = "unhealthy"
	MockHealthUnknown   MockHealth = "unknown"
)

type ManagedMock struct {
	ID     string
	Name   string
	Port   int
	Health MockHealth
}

const DefaultDebugPort = 8000
