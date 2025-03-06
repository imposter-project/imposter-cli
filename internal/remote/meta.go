package remote

import (
	"encoding/json"
	"fmt"
	"gatehill.io/imposter/internal/stringutil"
	"gatehill.io/imposter/internal/workspace"
	"os"
)

type RemoteMetadata struct {
	Workspace *workspace.Workspace
	Dir       string
	Config    map[string]string
}

func (m RemoteMetadata) CheckConfigKey(supported []string, key string) error {
	if stringutil.Contains(supported, key) {
		return nil
	}
	return fmt.Errorf("unsupported config key: %s", key)
}

func (m RemoteMetadata) SaveConfig() error {
	_, remoteFilePath, err := GetConfigPath(m.Dir, m.Workspace)
	if err != nil {
		return err
	}
	j, err := json.Marshal(m.Config)
	if err != nil {
		return fmt.Errorf("failed to marshall remote config for workspace: %s: %s", m.Workspace.Name, err)
	}
	err = os.WriteFile(remoteFilePath, j, 0644)
	if err != nil {
		return fmt.Errorf("failed to write remote config file: %s: %s", remoteFilePath, err)
	}
	return nil
}

func LoadConfig(dir string, w *workspace.Workspace, defaultsProvider func() *map[string]string) (c *map[string]string, err error) {
	exists, remoteFilePath, err := GetConfigPath(dir, w)
	if err != nil {
		return nil, err
	} else if !exists {
		return defaultsProvider(), nil
	}

	j, err := os.ReadFile(remoteFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load remote config file: %s: %s", remoteFilePath, err)
	}
	err = json.Unmarshal(j, &c)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall remote config file: %s: %s", remoteFilePath, err)
	}
	return c, nil
}

func CloneMap(orig *map[string]string) *map[string]string {
	clone := make(map[string]string, len(*orig))
	for k, v := range *orig {
		clone[k] = v
	}
	return &clone
}
