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

package cmd

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/imposter-project/imposter-cli/internal/config"
	"github.com/imposter-project/imposter-cli/internal/engine"
	"github.com/stretchr/testify/assert"
)

// fakeMockEngine is a minimal engine.MockEngine used to exercise the
// detach summary path; only GetID returns a meaningful value.
type fakeMockEngine struct{ id string }

func (f fakeMockEngine) Start(*sync.WaitGroup) bool                    { return true }
func (f fakeMockEngine) Stop(*sync.WaitGroup)                          {}
func (f fakeMockEngine) StopImmediately(*sync.WaitGroup)               {}
func (f fakeMockEngine) Restart(*sync.WaitGroup)                       {}
func (f fakeMockEngine) ListAllManaged() ([]engine.ManagedMock, error) { return nil, nil }
func (f fakeMockEngine) StopAllManaged() int                           { return 0 }
func (f fakeMockEngine) StopManaged(string) (bool, error)              { return false, nil }
func (f fakeMockEngine) GetVersionString() (string, error)             { return "", nil }
func (f fakeMockEngine) GetID() string                                 { return f.id }

func Test_applyDetachOptions(t *testing.T) {
	tmpHome := t.TempDir()
	config.DirPath = tmpHome
	defer func() { config.DirPath = "" }()

	t.Run("no detach leaves foreground mode and keeps auto-restart", func(t *testing.T) {
		opts := engine.StartOptions{Port: 8080}
		restart := applyDetachOptions(&opts, engine.EngineTypeJvmSingleJar, "", "", "", true)
		assert.Equal(t, engine.DetachNone, opts.Detach)
		assert.False(t, opts.IsDetached())
		assert.True(t, restart)
		assert.Empty(t, opts.DetachLog)
	})

	t.Run("detach=healthy waits for the healthcheck", func(t *testing.T) {
		opts := engine.StartOptions{Port: 8080}
		restart := applyDetachOptions(&opts, engine.EngineTypeJvmSingleJar, "healthy", "", "", true)
		assert.Equal(t, engine.DetachHealthy, opts.Detach)
		assert.False(t, restart, "auto-restart must be disabled when detached")
	})

	t.Run("detach=now returns immediately", func(t *testing.T) {
		opts := engine.StartOptions{Port: 8080}
		applyDetachOptions(&opts, engine.EngineTypeNative, "now", "", "", false)
		assert.Equal(t, engine.DetachNow, opts.Detach)
	})

	t.Run("process engine resolves default log path", func(t *testing.T) {
		opts := engine.StartOptions{Port: 1234}
		applyDetachOptions(&opts, engine.EngineTypeJvmSingleJar, "healthy", "", "", false)
		expected := filepath.Join(tmpHome, "logs", "imposter-1234.log")
		assert.Equal(t, expected, opts.DetachLog)
	})

	t.Run("explicit log file is honoured and made absolute", func(t *testing.T) {
		opts := engine.StartOptions{Port: 8080}
		applyDetachOptions(&opts, engine.EngineTypeNative, "healthy", "relative/mock.log", "", false)
		assert.True(t, filepath.IsAbs(opts.DetachLog))
		assert.Equal(t, "mock.log", filepath.Base(opts.DetachLog))
	})

	t.Run("docker engine does not set a detach log", func(t *testing.T) {
		opts := engine.StartOptions{Port: 8080}
		applyDetachOptions(&opts, engine.EngineTypeDockerCore, "healthy", "", "", false)
		assert.Equal(t, engine.DetachHealthy, opts.Detach)
		assert.Empty(t, opts.DetachLog)
	})

	t.Run("id file is stored when detached", func(t *testing.T) {
		opts := engine.StartOptions{Port: 8080}
		applyDetachOptions(&opts, engine.EngineTypeDockerCore, "healthy", "", "mock.id", false)
		assert.Equal(t, "mock.id", opts.DetachIdFile)
	})

	t.Run("id file is ignored without detach", func(t *testing.T) {
		opts := engine.StartOptions{Port: 8080}
		applyDetachOptions(&opts, engine.EngineTypeJvmSingleJar, "", "", "mock.id", false)
		assert.Empty(t, opts.DetachIdFile)
	})
}

func Test_writeMockIDFile(t *testing.T) {
	t.Run("writes plaintext id with no trailing newline", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "mock.id")
		err := writeMockIDFile(path, "abc123")
		assert.NoError(t, err)

		content, err := os.ReadFile(path)
		assert.NoError(t, err)
		assert.Equal(t, "abc123", string(content), "mock ID must be written as plaintext with no trailing newline")
	})

	t.Run("overwrites an existing file rather than appending", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "mock.id")
		assert.NoError(t, writeMockIDFile(path, "first"))
		assert.NoError(t, writeMockIDFile(path, "second"))

		content, err := os.ReadFile(path)
		assert.NoError(t, err)
		assert.Equal(t, "second", string(content), "re-running must replace the previous ID")
	})

	t.Run("returns an error when the path is not writable", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing-dir", "mock.id")
		assert.Error(t, writeMockIDFile(path, "abc123"))
	})
}

func Test_printDetachSummary(t *testing.T) {
	t.Run("writes the engine ID to the id-file when set", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "mock.id")
		printDetachSummary(fakeMockEngine{id: "engine-42"}, engine.StartOptions{Port: 8080, DetachIdFile: path})

		content, err := os.ReadFile(path)
		assert.NoError(t, err)
		assert.Equal(t, "engine-42", string(content))
	})

	t.Run("writes no file when id-file is unset", func(t *testing.T) {
		dir := t.TempDir()
		printDetachSummary(fakeMockEngine{id: "engine-42"}, engine.StartOptions{Port: 8080})

		entries, err := os.ReadDir(dir)
		assert.NoError(t, err)
		assert.Empty(t, entries, "no id-file should be created when the flag is unset")
	})

	t.Run("tolerates an unwritable id-file path", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing-dir", "mock.id")
		assert.NotPanics(t, func() {
			printDetachSummary(fakeMockEngine{id: "engine-42"}, engine.StartOptions{Port: 8080, DetachIdFile: path})
		})
		assert.NoFileExists(t, path)
	})
}
