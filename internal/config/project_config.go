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

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/imposter-project/imposter-cli/internal/fileutil"
)

// ProjectConfig holds the values used to scaffold a local project config file.
type ProjectConfig struct {
	Version string
	Plugins []string
}

// WriteProjectConfig writes a local project config file (imposter-project.yaml)
// to configDir using the canonical file name.
func WriteProjectConfig(configDir string, projectConfig ProjectConfig, forceOverwrite bool) {
	filePath := filepath.Join(configDir, LocalDirConfigFileName+".yaml")
	fileutil.MustNotExist(filePath, forceOverwrite)

	var b strings.Builder
	b.WriteString("# or pin to a particular version\n")
	fmt.Fprintf(&b, "version: %s\n", projectConfig.Version)
	b.WriteString("\n")
	b.WriteString("# See https://docs.imposter.sh/environment_variables/\n")
	b.WriteString("env:\n")
	b.WriteString("  IMPOSTER_LOG_LEVEL: DEBUG\n")

	if len(projectConfig.Plugins) > 0 {
		b.WriteString("\nplugins:\n")
		for _, p := range projectConfig.Plugins {
			fmt.Fprintf(&b, "  - %s\n", p)
		}
	}

	file, err := os.Create(filePath)
	if err != nil {
		logger.Fatal(err)
	}
	defer file.Close()

	_, err = file.WriteString(b.String())
	if err != nil {
		logger.Fatal(err)
	}

	logger.Infof("wrote project config: %v", filePath)
}
