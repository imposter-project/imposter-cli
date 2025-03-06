package jvm

import (
	"fmt"
	library2 "gatehill.io/imposter/internal/library"
	"os"
	"os/exec"
	"path/filepath"

	"gatehill.io/imposter/internal/engine"
	"github.com/spf13/viper"
)

type SingleJarProvider struct {
	JvmProviderOptions
	jarPath string
}

const binCacheDir = ".imposter/engines/jvm"

var singleJarInitialised = false

// EnableEngine is a convenience function that delegates to EnableSingleJarEngine.
func EnableEngine() {
	EnableSingleJarEngine()
}

func EnableSingleJarEngine() {
	if !singleJarInitialised {
		singleJarInitialised = true

		engine.RegisterLibrary(engine.EngineTypeJvmSingleJar, func() engine.EngineLibrary {
			return getSingleJarLibrary()
		})
		engine.RegisterEngine(engine.EngineTypeJvmSingleJar, func(configDir string, startOptions engine.StartOptions) engine.MockEngine {
			provider := newSingleJarProvider(startOptions.Version)
			return buildEngine(configDir, &provider, startOptions)
		})
	}
}

func getSingleJarLibrary() *JvmEngineLibrary {
	return &JvmEngineLibrary{engineType: engine.EngineTypeJvmSingleJar}
}

func newSingleJarProvider(version string) JvmProvider {
	return &SingleJarProvider{
		JvmProviderOptions: JvmProviderOptions{
			EngineMetadata: engine.EngineMetadata{
				EngineType: engine.EngineTypeJvmSingleJar,
				Version:    version,
			},
		},
	}
}

func (p *SingleJarProvider) GetStartCommand(args []string, env []string) *exec.Cmd {
	if p.javaCmd == "" {
		javaCmd, err := GetJavaCmdPath()
		if err != nil {
			logger.Fatal(err)
		}
		p.javaCmd = javaCmd
	}
	if !p.Satisfied() {
		if err := p.Provide(engine.PullIfNotPresent); err != nil {
			logger.Fatal(err)
		}
	}
	allArgs := append(
		[]string{"-jar", p.jarPath},
		args...,
	)
	command := exec.Command(p.javaCmd, allArgs...)
	command.Env = env
	return command
}

func (p *SingleJarProvider) Provide(policy engine.PullPolicy) error {
	jarPath, err := ensureBinary(p.Version, policy)
	if err != nil {
		return err
	}
	p.jarPath = jarPath
	return nil
}

func (p *SingleJarProvider) Satisfied() bool {
	return p.jarPath != ""
}

func ensureBinary(version string, policy engine.PullPolicy) (string, error) {
	if envJarFile := viper.GetString("jvm.jarFile"); envJarFile != "" {
		if _, err := os.Stat(envJarFile); err != nil {
			return "", fmt.Errorf("could not stat JAR file: %v: %v", envJarFile, err)
		}
		logger.Debugf("using JAR file: %v", envJarFile)
		return envJarFile, nil
	}
	return checkOrDownloadBinary(version, policy)
}

func checkOrDownloadBinary(version string, policy engine.PullPolicy) (string, error) {
	binCachePath, err := ensureBinCache()
	if err != nil {
		logger.Fatal(err)
	}

	versionedBinDir := filepath.Join(binCachePath, version)
	if err := os.MkdirAll(versionedBinDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create versioned bin directory: %v", err)
	}

	binFilePath := filepath.Join(versionedBinDir, "imposter.jar")
	if policy == engine.PullSkip {
		return binFilePath, nil
	}

	if policy == engine.PullIfNotPresent {
		if _, err = os.Stat(binFilePath); err != nil {
			if !os.IsNotExist(err) {
				return "", fmt.Errorf("failed to stat: %v: %v", binFilePath, err)
			}
		} else {
			logger.Debugf("engine binary '%v' already present", version)
			logger.Tracef("binary for version %v found at: %v", version, binFilePath)
			return binFilePath, nil
		}
	}

	if err := downloadBinary(binFilePath, version); err != nil {
		return "", fmt.Errorf("failed to fetch binary: %v", err)
	}
	logger.Tracef("using imposter at: %v", binFilePath)
	return binFilePath, nil
}

func ensureBinCache() (string, error) {
	return library2.EnsureDirUsingConfig("jvm.binCache", binCacheDir)
}

func downloadBinary(localPath string, version string) error {
	fallbackRemoteFileName := fmt.Sprintf("imposter-%v.jar", version)
	return library2.DownloadBinaryWithFallback(localPath, "imposter.jar", version, fallbackRemoteFileName)
}
