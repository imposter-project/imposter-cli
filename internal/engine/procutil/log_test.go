package procutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_OpenDetachLog(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "imposter-8080.log")

	f, err := OpenDetachLog(path)
	require.NoError(t, err)
	_, err = f.WriteString("first\n")
	require.NoError(t, err)
	require.NoError(t, f.Close())

	// reopening must append, not truncate
	f2, err := OpenDetachLog(path)
	require.NoError(t, err)
	_, err = f2.WriteString("second\n")
	require.NoError(t, err)
	require.NoError(t, f2.Close())

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "first\nsecond\n", string(content))
}
