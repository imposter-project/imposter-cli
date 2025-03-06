package jvm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gatehill.io/imposter/internal/engine"
)

type JvmEngineLibrary struct {
	engineType engine.EngineType
}

func (JvmEngineLibrary) CheckPrereqs() (bool, []string) {
	var msgs []string
	javaCmdPath, err := GetJavaCmdPath()
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("❌ Failed to find JVM installation: %v", err))
		return false, msgs
	}
	msgs = append(msgs, fmt.Sprintf("✅ Found JVM installation: %v", javaCmdPath))

	cmd := exec.Command(javaCmdPath, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("❌ Failed to determine java version: %v", err))
		return false, msgs
	}
	msgs = append(msgs, fmt.Sprintf("✅ Java version installed: %v", string(output)))

	return true, msgs
}

func (JvmEngineLibrary) List() ([]engine.EngineMetadata, error) {
	binCachePath, err := ensureBinCache()
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(binCachePath)
	if err != nil {
		return nil, fmt.Errorf("error reading binary cache directory: %v: %v", binCachePath, err)
	}
	var available []engine.EngineMetadata
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		// Check if the versioned directory contains a JAR file
		versionDir := filepath.Join(binCachePath, file.Name())
		jarPath := filepath.Join(versionDir, "imposter.jar")
		if _, err := os.Stat(jarPath); err == nil {
			available = append(available, engine.EngineMetadata{
				EngineType: engine.EngineTypeJvmSingleJar,
				Version:    file.Name(),
			})
		}
	}
	return available, nil
}

func (j JvmEngineLibrary) GetProvider(version string) engine.Provider {
	switch j.engineType {
	case engine.EngineTypeJvmSingleJar:
		return newSingleJarProvider(version)
	case engine.EngineTypeJvmUnpacked:
		return newUnpackedDistroProvider(version)
	default:
		panic(fmt.Errorf("unsupported engine type: %s for JVM library", j.engineType))
	}
}

func (j JvmEngineLibrary) IsSealedDistro() bool {
	switch j.engineType {
	case engine.EngineTypeJvmSingleJar:
		return false
	case engine.EngineTypeJvmUnpacked:
		return true
	default:
		panic(fmt.Errorf("unsupported engine type: %s for JVM library", j.engineType))
	}
}

func (j JvmEngineLibrary) ShouldEnsurePlugins() bool {
	return !j.IsSealedDistro()
}
