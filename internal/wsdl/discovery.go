package wsdl

import (
	"gatehill.io/imposter/internal/fileutil"
	"path/filepath"
)

// DiscoverWSDLFiles finds WSDL files within the given directory.
// It returns fully qualified paths to the files discovered.
func DiscoverWSDLFiles(configDir string) []string {
	var wsdlFiles []string

	candidates := fileutil.FindFilesWithExtension(configDir, ".wsdl")
	for _, candidate := range candidates {
		fullyQualifiedPath := filepath.Join(configDir, candidate)
		wsdlFiles = append(wsdlFiles, fullyQualifiedPath)
	}

	return wsdlFiles
}
