package impostermodel

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/imposter-project/imposter-cli/internal/fileutil"
)

type DotImposterConfig struct {
	Version string
	Plugins []string
}

func writeDotImposterYaml(configDir string, dotConfig DotImposterConfig, forceOverwrite bool) {
	filePath := filepath.Join(configDir, ".imposter.yaml")
	fileutil.MustNotExist(filePath, forceOverwrite)

	var b strings.Builder
	b.WriteString("# or pin to a particular version\n")
	fmt.Fprintf(&b, "version: %s\n", dotConfig.Version)
	b.WriteString("\n")
	b.WriteString("# See https://docs.imposter.sh/environment_variables/\n")
	b.WriteString("env:\n")
	b.WriteString("  IMPOSTER_LOG_LEVEL: DEBUG\n")

	if len(dotConfig.Plugins) > 0 {
		b.WriteString("\nplugins:\n")
		for _, p := range dotConfig.Plugins {
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

	logger.Infof("wrote Imposter config: %v", filePath)
}
