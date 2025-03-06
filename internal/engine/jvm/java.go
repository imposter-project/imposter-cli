package jvm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// GetJavaCmdPath finds the best candidate for the 'java' command, searching
// the environment as well as using well-known OS-specific mechanisms.
func GetJavaCmdPath() (string, error) {
	// search for 'java' in the PATH
	javaPath, err := exec.LookPath("java")
	if err != nil {
		logger.Tracef("could not find 'java' in PATH: %s", err)
	}

	var binaryPathSuffix string
	if runtime.GOOS == "Windows" {
		binaryPathSuffix = ".exe"
	} else {
		binaryPathSuffix = ""
	}

	if javaPath == "" {
		// check JAVA_HOME environment variable
		if javaHomeEnv, found := os.LookupEnv("JAVA_HOME"); found {
			javaPath = filepath.Join(javaHomeEnv, "/bin/java"+binaryPathSuffix)
		}
	}

	if javaPath == "" && runtime.GOOS == "darwin" {
		command, stdout := exec.Command("/usr/libexec/java_home"), new(strings.Builder)
		command.Stdout = stdout
		err := command.Run()
		if err != nil {
			return "", fmt.Errorf("error determining JAVA_HOME: %s", err)
		}
		if command.ProcessState.Success() {
			javaPath = filepath.Join(strings.TrimSpace(stdout.String()), "/bin/java"+binaryPathSuffix)
		} else {
			return "", fmt.Errorf("failed to determine JAVA_HOME using libexec")
		}
	}

	if javaPath == "" {
		return "", fmt.Errorf("failed to determine Java path - consider setting JAVA_HOME or updating PATH")
	}

	logger.Tracef("using java: %v", javaPath)
	return javaPath, nil
}
