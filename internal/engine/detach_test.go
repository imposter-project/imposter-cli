package engine

import (
	"path/filepath"
	"testing"

	"github.com/imposter-project/imposter-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_DefaultDetachLogPath(t *testing.T) {
	tmpHome := t.TempDir()
	config.DirPath = tmpHome
	defer func() { config.DirPath = "" }()

	path, err := DefaultDetachLogPath(8081)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpHome, "logs", "imposter-8081.log"), path)
}

func Test_StartOptions_IsDetached(t *testing.T) {
	assert.False(t, StartOptions{Detach: DetachNone}.IsDetached())
	assert.True(t, StartOptions{Detach: DetachNow}.IsDetached())
	assert.True(t, StartOptions{Detach: DetachHealthy}.IsDetached())
}
