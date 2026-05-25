package awslambda

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/imposter-project/imposter-cli/internal/compression"
	"github.com/imposter-project/imposter-cli/internal/engine"
	library2 "github.com/imposter-project/imposter-cli/internal/library"
	"github.com/spf13/viper"
)

// DefaultLambdaArch is the architecture assumed when a caller does not
// specify one. It matches AWS Lambda's own default architecture (x86_64,
// expressed as the Go arch name "amd64") so unconfigured bundles still
// produce a deployable artefact.
const DefaultLambdaArch = "amd64"

// lambdaBinarySpec materialises a Lambda-ready zip into the local cache for
// a given engine flavour. cacheFile names the on-disk artefact; assemble
// fetches and (where necessary) converts the upstream release into that zip.
type lambdaBinarySpec struct {
	cacheFile func(arch, version string) string
	assemble  func(cachePath, version, arch string) error
}

// jvmLambdaSpec ships a pre-built Lambda zip from imposter-jvm-engine that
// is already arch-agnostic and Lambda-ready, so assembly is a direct
// download.
var jvmLambdaSpec = lambdaBinarySpec{
	cacheFile: func(_, version string) string {
		return fmt.Sprintf("imposter-awslambda-%s.zip", version)
	},
	assemble: func(cachePath, version, _ string) error {
		dc := library2.NewDownloadConfig(
			"https://github.com/imposter-project/imposter-jvm-engine/releases/latest/download",
			"https://github.com/imposter-project/imposter-jvm-engine/releases/download/v%v",
			false,
		)
		return library2.DownloadBinary(dc, cachePath, "imposter-awslambda.zip", version)
	},
}

// nativeLambdaSpec consumes the per-arch imposter-go release tarball and
// repackages the contained binary as a Lambda custom-runtime zip whose
// single entry is the executable "bootstrap".
var nativeLambdaSpec = lambdaBinarySpec{
	cacheFile: func(arch, version string) string {
		return fmt.Sprintf("imposter-go-awslambda-%s-%s.zip", arch, version)
	},
	assemble: assembleNativeLambdaZip,
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

		if err := spec.assemble(binFilePath, version, arch); err != nil {
			return "", fmt.Errorf("failed to fetch lambda binary: %v", err)
		}
	}
	logger.Tracef("using lambda binary at: %v", binFilePath)
	return binFilePath, nil
}

// assembleNativeLambdaZip downloads the imposter-go linux release tarball
// for the requested architecture, extracts the imposter-go binary, and
// writes a Lambda-ready zip to cachePath containing a single executable
// "bootstrap" entry (as required by the provided.al2023 custom runtime).
func assembleNativeLambdaZip(cachePath, version, arch string) error {
	dc := library2.NewDownloadConfig(
		"https://github.com/imposter-project/imposter-go/releases/latest/download",
		"https://github.com/imposter-project/imposter-go/releases/download/v%v",
		false,
	)
	remoteFile := fmt.Sprintf("imposter-go_linux_%s.tar.gz", arch)

	tempDir, err := os.MkdirTemp("", "imposter-go-lambda-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tarballPath := filepath.Join(tempDir, remoteFile)
	if err := library2.DownloadBinary(dc, tarballPath, remoteFile, version); err != nil {
		return err
	}

	binaryPath, err := extractGoBinary(tarballPath, tempDir)
	if err != nil {
		return fmt.Errorf("failed to extract imposter-go binary from %s: %v", tarballPath, err)
	}

	if err := writeBootstrapZip(binaryPath, cachePath); err != nil {
		return fmt.Errorf("failed to write lambda zip %s: %v", cachePath, err)
	}
	return nil
}

// extractGoBinary extracts the imposter-go release tarball and returns the
// path to the extracted "imposter-go" binary. The release archive currently
// also contains README/CHANGELOG files which are ignored.
func extractGoBinary(tarballPath, destDir string) (string, error) {
	if err := compression.ExtractTarGz(tarballPath, destDir); err != nil {
		return "", err
	}
	binaryPath := filepath.Join(destDir, "imposter-go")
	if _, err := os.Stat(binaryPath); err != nil {
		return "", fmt.Errorf("imposter-go binary not found in archive: %v", err)
	}
	return binaryPath, nil
}

// writeBootstrapZip writes a zip containing a single "bootstrap" entry with
// the contents of binaryPath. The entry is marked executable so AWS
// Lambda's provided.al2023 runtime will run it.
func writeBootstrapZip(binaryPath, zipPath string) error {
	src, err := os.Open(binaryPath)
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	defer zw.Close()

	header := &zip.FileHeader{
		Name:   "bootstrap",
		Method: zip.Deflate,
	}
	header.SetMode(0755)

	w, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, src); err != nil {
		return err
	}
	return nil
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
