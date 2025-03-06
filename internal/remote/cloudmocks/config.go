package cloudmocks

import (
	"fmt"
	"gatehill.io/imposter/internal/logging"
	"gatehill.io/imposter/internal/prefs"
	remote2 "gatehill.io/imposter/internal/remote"
	"gatehill.io/imposter/internal/workspace"
	"net/url"
	"strings"
)

const remoteType = "cloudmocks"
const defaultUrl = "https://api.mocks.cloud"

const configKeyMockId = "mockId"
const configKeyToken = "token"
const configKeyUrl = "url"

type CloudMocksRemote struct {
	remote2.RemoteMetadata
}

var configKeys = []string{
	configKeyMockId,
	configKeyToken,
	configKeyUrl,
}

var logger = logging.GetLogger()

func Register() {
	remote2.Register(remoteType, func(dir string, workspace *workspace.Workspace) (remote2.Remote, error) {
		return Load(dir, workspace)
	})
}

func Load(dir string, w *workspace.Workspace) (CloudMocksRemote, error) {
	c, err := remote2.LoadConfig(dir, w, func() *map[string]string {
		return &map[string]string{
			configKeyUrl: defaultUrl,
		}
	})
	if err != nil {
		return CloudMocksRemote{}, err
	}

	r := CloudMocksRemote{
		remote2.RemoteMetadata{
			Workspace: w,
			Dir:       dir,
			Config:    *c,
		},
	}
	return r, nil
}

func (CloudMocksRemote) GetType() string {
	return remoteType
}

func (CloudMocksRemote) GetConfigKeys() []string {
	return configKeys
}

func (m CloudMocksRemote) SetConfigValue(key string, value string) error {
	if err := m.CheckConfigKey(m.GetConfigKeys(), key); err != nil {
		return err
	}

	switch key {
	case configKeyUrl:
		value = strings.TrimSuffix(value, "/")
		if _, err := url.Parse(value); err != nil {
			return fmt.Errorf("failed to parse URL: %s: %s", value, err)
		}
		break

	case configKeyToken:
		token := value
		value = ""
		if err := m.setToken(token); err != nil {
			return err
		}
		// do not persist token to config
		return nil
	}
	m.Config[key] = value
	return m.SaveConfig()
}

func (m CloudMocksRemote) GetConfig() (*map[string]string, error) {
	cfg := *remote2.CloneMap(&m.Config)
	token, err := m.getObfuscatedToken()
	if err != nil {
		return nil, err
	}
	cfg["token"] = token
	return &cfg, nil
}

func (m CloudMocksRemote) setToken(token string) error {
	return getCredsPrefs().WriteProperty(m.Config[configKeyUrl], token)
}

func (m CloudMocksRemote) getCleartextToken() (string, error) {
	cleartext, err := getCredsPrefs().ReadPropertyString(m.Config[configKeyUrl])
	if err != nil {
		return "", err
	}
	return cleartext, nil
}

func (m CloudMocksRemote) getObfuscatedToken() (string, error) {
	cleartext, err := m.getCleartextToken()
	if err != nil {
		return "", err
	} else if cleartext == "" {
		return "", nil
	}
	obfuscated := strings.Repeat("*", 8) + cleartext[len(cleartext)-4:]
	return obfuscated, nil
}

func (m CloudMocksRemote) GetStatus() (*remote2.Status, error) {
	s, err := m.getStatus()
	if err != nil {
		return nil, err
	}
	status := remote2.Status{
		Status:       s.Status,
		LastModified: int64(s.LastModified),
	}
	return &status, nil
}

func getCredsPrefs() prefs.Prefs {
	return prefs.Load("credentials.json")
}
