package procutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// OpenDetachLog opens (creating directories and file as needed) the log
// file a detached process engine writes its stdout/stderr to. The file is
// opened in append mode so restarts do not truncate prior output.
func OpenDetachLog(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create log directory for %s: %w", path, err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open detach log %s: %w", path, err)
	}
	return f, nil
}
