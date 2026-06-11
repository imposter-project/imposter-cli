package protobuf

import (
	"github.com/imposter-project/imposter-cli/internal/fileutil"
	"path/filepath"
)

// DiscoverProtoFiles finds protobuf (.proto) files within the given directory.
// It returns fully qualified paths to the files discovered.
func DiscoverProtoFiles(configDir string) []string {
	var protoFiles []string

	candidates := fileutil.FindFilesWithExtension(configDir, ".proto")
	for _, candidate := range candidates {
		fullyQualifiedPath := filepath.Join(configDir, candidate)
		protoFiles = append(protoFiles, fullyQualifiedPath)
	}

	return protoFiles
}
