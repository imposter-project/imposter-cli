package library

import (
	"fmt"
	"gatehill.io/imposter/internal/logging"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

var logger = logging.GetLogger()

func EnsureDirUsingConfig(settingsKey string, homeSubDirPath string) (string, error) {
	dirPath, err := GetDirPath(settingsKey, homeSubDirPath)
	if err != nil {
		return "", err
	}
	if err = EnsureDir(dirPath); err != nil {
		return "", err
	}
	logger.Tracef("ensured directory: %v", dirPath)
	return dirPath, nil
}

func GetDirPath(settingsKey string, homeSubDirPath string) (string, error) {
	if envDirPath := viper.GetString(settingsKey); envDirPath != "" {
		return envDirPath, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %v", err)
	}
	return filepath.Join(homeDir, homeSubDirPath), nil
}

func EnsureDir(dirPath string) error {
	if info, err := os.Stat(dirPath); err != nil {
		if os.IsNotExist(err) {
			logger.Tracef("creating directory: %v", dirPath)
			err := os.MkdirAll(dirPath, 0700)
			if err != nil {
				return fmt.Errorf("failed to create directory: %v: %v", dirPath, err)
			}
		} else {
			return fmt.Errorf("failed to stat: %v: %v", dirPath, err)
		}
	} else if !info.IsDir() {
		return fmt.Errorf("path: %s is not a directory", dirPath)
	}
	return nil
}
