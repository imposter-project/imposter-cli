package awslambda

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/imposter-project/imposter-cli/internal/engine"
	library2 "github.com/imposter-project/imposter-cli/internal/library"
	"github.com/spf13/viper"
)

// DefaultLambdaArch is the architecture assumed when a caller does not
// specify one. It matches AWS Lambda's own default architecture (x86_64,
// expressed as the Go arch name "amd64") so unconfigured bundles still
// produce a deployable artefact.
const DefaultLambdaArch = "amd64"

// lambdaBinarySpec describes the source and local naming for an AWS Lambda
// deployment binary. The JVM-engine flavour ships a single arch-agnostic zip
// from imposter-jvm-engine; the native (imposter-go) flavour ships separate
// per-architecture zips from imposter-go.
type lambdaBinarySpec struct {
	downloadConfig library2.DownloadConfig
	remoteFile     func(arch string) string
	cacheFile      func(arch, version string) string
}

var jvmLambdaSpec = lambdaBinarySpec{
	downloadConfig: library2.NewDownloadConfig(
		"https://github.com/imposter-project/imposter-jvm-engine/releases/latest/download",
		"https://github.com/imposter-project/imposter-jvm-engine/releases/download/v%v",
		false,
	),
	remoteFile: func(string) string { return "imposter-awslambda.zip" },
	cacheFile: func(_, version string) string {
		return fmt.Sprintf("imposter-awslambda-%s.zip", version)
	},
}

var nativeLambdaSpec = lambdaBinarySpec{
	downloadConfig: library2.NewDownloadConfig(
		"https://github.com/imposter-project/imposter-go/releases/latest/download",
		"https://github.com/imposter-project/imposter-go/releases/download/v%v",
		false,
	),
	remoteFile: func(arch string) string {
		return fmt.Sprintf("imposter-awslambda_%s.zip", arch)
	},
	cacheFile: func(arch, version string) string {
		return fmt.Sprintf("imposter-go-awslambda-%s-%s.zip", arch, version)
	},
}

// specForVersion picks the AWS Lambda binary spec for the given engine
// version. 5.x and later use the native (imposter-go) flavour; everything
// else — including the empty/"latest" alias and unparseable values — falls
// back to the JVM flavour, matching the project-wide default engine.
func specForVersion(version string) lambdaBinarySpec {
	if engine.DeriveEngineTypeFromVersion(version) == engine.EngineTypeNative {
		return nativeLambdaSpec
	}
	return jvmLambdaSpec
}

func checkOrDownloadBinary(version string, arch string) (string, error) {
	if arch == "" {
		arch = DefaultLambdaArch
	}
	binFilePath := viper.GetString("lambda.binary")
	if binFilePath == "" {
		spec := specForVersion(version)

		binCachePath, err := ensureBinCache()
		if err != nil {
			logger.Fatal(err)
		}

		binFilePath = filepath.Join(binCachePath, spec.cacheFile(arch, version))

		if _, err := os.Stat(binFilePath); err != nil {
			if !os.IsNotExist(err) {
				return "", fmt.Errorf("failed to stat: %v: %v", binFilePath, err)
			}
		} else {
			logger.Debugf("lambda binary '%v' already present", version)
			logger.Tracef("lambda binary for version %v found at: %v", version, binFilePath)
			return binFilePath, nil
		}

		if err := library2.DownloadBinary(spec.downloadConfig, binFilePath, spec.remoteFile(arch), version); err != nil {
			return "", fmt.Errorf("failed to fetch lambda binary: %v", err)
		}
	}
	logger.Tracef("using lambda binary at: %v", binFilePath)
	return binFilePath, nil
}

func ensureBinCache() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %v", err)
	}
	dirPath := filepath.Join(homeDir, ".imposter/awslambda")
	if err = library2.EnsureDir(dirPath); err != nil {
		return "", err
	}
	logger.Tracef("ensured directory: %v", dirPath)
	return dirPath, nil
}
