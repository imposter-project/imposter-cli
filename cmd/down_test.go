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
	"testing"

	"github.com/imposter-project/imposter-cli/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_stopEngine(t *testing.T) {
	tests := []struct {
		name       string
		engineType engine.EngineType
	}{
		{
			name:       "stop docker engine",
			engineType: engine.EngineTypeDockerCore,
		},
		{
			name:       "stop jvm engine",
			engineType: engine.EngineTypeJvmSingleJar,
		},
		{
			name:       "stop native engine",
			engineType: engine.EngineTypeNative,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stopped, err := stopEngine(tt.engineType)
			require.NoError(t, err)
			require.GreaterOrEqual(t, stopped, 0)
		})
	}
}

func Test_downCmd_requires_id_or_all(t *testing.T) {
	// bare `imposter down` is fatal — capture via runWithRecovery
	rootCmd.SetArgs([]string{"down"})
	err := runWithRecovery(func() {
		_ = rootCmd.Execute()
	})
	require.Error(t, err, "should fail without an ID or --all")
}

func Test_downCmd_rejects_id_with_all(t *testing.T) {
	resetDownFlags()
	rootCmd.SetArgs([]string{"down", "--all", "abc123"})
	err := runWithRecovery(func() {
		_ = rootCmd.Execute()
	})
	require.Error(t, err, "should reject ID together with --all")
}

func Test_downCmd_rejects_id_file_with_all(t *testing.T) {
	resetDownFlags()
	idFile := writeTempIDFile(t, "abc123")
	rootCmd.SetArgs([]string{"down", "--all", "--id-file", idFile})
	err := runWithRecovery(func() {
		_ = rootCmd.Execute()
	})
	require.Error(t, err, "should reject --id-file together with --all")
}

func Test_downCmd_rejects_id_file_with_id_arg(t *testing.T) {
	resetDownFlags()
	idFile := writeTempIDFile(t, "abc123")
	rootCmd.SetArgs([]string{"down", "abc123", "--id-file", idFile})
	err := runWithRecovery(func() {
		_ = rootCmd.Execute()
	})
	require.Error(t, err, "should reject a mock ID together with --id-file")
}

func Test_downCmd_id_file_missing_is_fatal(t *testing.T) {
	resetDownFlags()
	rootCmd.SetArgs([]string{"down", "--id-file", filepath.Join(t.TempDir(), "absent.id")})
	err := runWithRecovery(func() {
		_ = rootCmd.Execute()
	})
	require.Error(t, err, "should fail when the id file cannot be read")
}

func Test_readMockIDFile(t *testing.T) {
	t.Run("reads and trims surrounding whitespace", func(t *testing.T) {
		path := writeTempIDFile(t, "  abc123\n")
		id, err := readMockIDFile(path)
		require.NoError(t, err)
		assert.Equal(t, "abc123", id)
	})

	t.Run("errors when the file is missing", func(t *testing.T) {
		_, err := readMockIDFile(filepath.Join(t.TempDir(), "absent.id"))
		assert.Error(t, err)
	})

	t.Run("errors when the file holds no ID", func(t *testing.T) {
		path := writeTempIDFile(t, "   \n\t")
		_, err := readMockIDFile(path)
		assert.Error(t, err)
	})
}

func resetDownFlags() {
	downFlags.all = false
	downFlags.idFile = ""
}

func writeTempIDFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "mock.id")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	return path
}

func Test_stopMockByID_unknown(t *testing.T) {
	err := runWithRecovery(func() {
		stopMockByID("definitely-not-a-real-mock-id")
	})
	require.Error(t, err, "should fail when no engine has a mock with the given id")
}
