package workspace

import (
	"encoding/json"
	"fmt"
	"gatehill.io/imposter/internal/library"
	"os"
	"path/filepath"
)

type Workspace struct {
	Name       string `json:"name"`
	RemoteType string `json:"remoteType"`
}

type Metadata struct {
	Workspaces []*Workspace `json:"workspaces"`
	Active     string       `json:"active"`
}

const metaDirName = ".imposter"

func createOrLoadMetadata(dir string) (m *Metadata, err error) {
	metaFilePath, err := getMetaFilePath(dir)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(metaFilePath); err != nil {
		if os.IsNotExist(err) {
			logger.Trace("creating empty workspace metadata")
			m = &Metadata{}
		} else {
			return nil, fmt.Errorf("failed to stat workspace file: %s: %s", metaFilePath, err)
		}
	} else {
		j, err := os.ReadFile(metaFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read workspace file: %s: %s", metaFilePath, err)
		}
		err = json.Unmarshal(j, &m)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshall workspace file: %s: %s", metaFilePath, err)
		}
	}
	return m, nil
}

func EnsureMetadataDir(dir string) (string, error) {
	metaDir := filepath.Join(dir, metaDirName)
	if err := library.EnsureDir(metaDir); err != nil {
		return "", fmt.Errorf("failed to ensure workspace metadata directory exists: %s: %s", metaDir, err)
	}
	return metaDir, nil
}

func SaveMetadata(dir string, m *Metadata) error {
	metaFilePath, err := getMetaFilePath(dir)
	if err != nil {
		return err
	}
	j, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshall metadata: %s", err)
	}
	err = os.WriteFile(metaFilePath, j, 0644)
	if err != nil {
		return fmt.Errorf("failed to save metadata to: %s: %s", metaFilePath, err)
	}
	return nil
}

func getMetaFilePath(dir string) (string, error) {
	metaDir, err := EnsureMetadataDir(dir)
	if err != nil {
		return "", err
	}
	metaFilePath := filepath.Join(metaDir, "workspaces.json")
	return metaFilePath, nil
}
