package mockscloud

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/imposter-project/imposter-cli/internal/logging"
	"github.com/imposter-project/imposter-cli/internal/prefs"
	remote2 "github.com/imposter-project/imposter-cli/internal/remote"
	"github.com/imposter-project/imposter-cli/internal/workspace"
)

const remoteType = "mockscloud"
const defaultUrl = "https://api.mocks.cloud"

const configKeyMockId = "mockId"
const configKeyToken = "token"
const configKeyUrl = "url"

type MocksCloudRemote struct {
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

func Load(dir string, w *workspace.Workspace) (MocksCloudRemote, error) {
	c, err := remote2.LoadConfig(dir, w, func() *map[string]string {
		return &map[string]string{
			configKeyUrl: defaultUrl,
		}
	})
	if err != nil {
		return MocksCloudRemote{}, err
	}

	r := MocksCloudRemote{
		remote2.RemoteMetadata{
			Workspace: w,
			Dir:       dir,
			Config:    *c,
		},
	}
	return r, nil
}

func (MocksCloudRemote) GetType() string {
	return remoteType
}

func (MocksCloudRemote) GetConfigKeys() []string {
	return configKeys
}

func (m MocksCloudRemote) SetConfigValue(key string, value string) error {
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

func (m MocksCloudRemote) GetConfig() (*map[string]string, error) {
	cfg := *remote2.CloneMap(&m.Config)
	token, err := m.getObfuscatedToken()
	if err != nil {
		return nil, err
	}
	cfg["token"] = token
	return &cfg, nil
}

func (m MocksCloudRemote) setToken(token string) error {
	return getCredsPrefs().WriteProperty(m.Config[configKeyUrl], token)
}

func (m MocksCloudRemote) getCleartextToken() (string, error) {
	cleartext, err := getCredsPrefs().ReadPropertyString(m.Config[configKeyUrl])
	if err != nil {
		return "", err
	}
	return cleartext, nil
}

func (m MocksCloudRemote) getObfuscatedToken() (string, error) {
	cleartext, err := m.getCleartextToken()
	if err != nil {
		return "", err
	} else if cleartext == "" {
		return "", nil
	}
	obfuscated := strings.Repeat("*", 8) + cleartext[len(cleartext)-4:]
	return obfuscated, nil
}

func (m MocksCloudRemote) GetStatus() (*remote2.Status, error) {
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
