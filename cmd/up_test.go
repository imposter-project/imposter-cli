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
	"path/filepath"
	"testing"

	"github.com/imposter-project/imposter-cli/internal/config"
	"github.com/imposter-project/imposter-cli/internal/engine"
	"github.com/stretchr/testify/assert"
)

func Test_applyDetachOptions(t *testing.T) {
	tmpHome := t.TempDir()
	config.DirPath = tmpHome
	defer func() { config.DirPath = "" }()

	t.Run("no detach leaves foreground mode and keeps auto-restart", func(t *testing.T) {
		opts := engine.StartOptions{Port: 8080}
		restart := applyDetachOptions(&opts, engine.EngineTypeJvmSingleJar, false, false, "", true)
		assert.Equal(t, engine.DetachNone, opts.Detach)
		assert.False(t, opts.IsDetached())
		assert.True(t, restart)
		assert.Empty(t, opts.DetachLog)
	})

	t.Run("detach defaults to await-healthy", func(t *testing.T) {
		opts := engine.StartOptions{Port: 8080}
		restart := applyDetachOptions(&opts, engine.EngineTypeJvmSingleJar, true, false, "", true)
		assert.Equal(t, engine.DetachHealthy, opts.Detach)
		assert.False(t, restart, "auto-restart must be disabled when detached")
	})

	t.Run("detach with no-await returns immediately", func(t *testing.T) {
		opts := engine.StartOptions{Port: 8080}
		applyDetachOptions(&opts, engine.EngineTypeNative, true, true, "", false)
		assert.Equal(t, engine.DetachNow, opts.Detach)
	})

	t.Run("process engine resolves default log path", func(t *testing.T) {
		opts := engine.StartOptions{Port: 1234}
		applyDetachOptions(&opts, engine.EngineTypeJvmSingleJar, true, false, "", false)
		expected := filepath.Join(tmpHome, "logs", "imposter-1234.log")
		assert.Equal(t, expected, opts.DetachLog)
	})

	t.Run("explicit log file is honoured and made absolute", func(t *testing.T) {
		opts := engine.StartOptions{Port: 8080}
		applyDetachOptions(&opts, engine.EngineTypeNative, true, false, "relative/mock.log", false)
		assert.True(t, filepath.IsAbs(opts.DetachLog))
		assert.Equal(t, "mock.log", filepath.Base(opts.DetachLog))
	})

	t.Run("docker engine does not set a detach log", func(t *testing.T) {
		opts := engine.StartOptions{Port: 8080}
		applyDetachOptions(&opts, engine.EngineTypeDockerCore, true, false, "", false)
		assert.Equal(t, engine.DetachHealthy, opts.Detach)
		assert.Empty(t, opts.DetachLog)
	})
}
