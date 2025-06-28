package compression

import (
	"fmt"
	"gatehill.io/imposter/internal/logging"
	"strings"
)

var logger = logging.GetLogger()

// IsArchiveFileExtension checks if the given file name has an archive file extension.
func IsArchiveFileExtension(name string) bool {
	archiveExtensions := []string{".zip", ".tar.gz"}
	for _, ext := range archiveExtensions {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}

// ExtractArchive extracts an archive file (zip or tar.gz) to the specified destination directory.
func ExtractArchive(archiveFilePath string, destinationDir string) error {
	if strings.HasSuffix(archiveFilePath, ".zip") {
		return ExtractZip(archiveFilePath, destinationDir)
	} else if strings.HasSuffix(archiveFilePath, ".tar.gz") {
		return ExtractTarGz(archiveFilePath, destinationDir)
	}
	return fmt.Errorf("unsupported archive format: %s", archiveFilePath)
}
