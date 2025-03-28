package docker

import (
	"gatehill.io/imposter/internal/engine"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"testing"
)

func TestImages_getImageRepo_defaultReg(t *testing.T) {
	logger.SetLevel(logrus.TraceLevel)
	expected := map[string]string{
		"docker":            "outofcoffee/imposter",
		"docker-all":        "outofcoffee/imposter-all",
		"docker-distroless": "outofcoffee/imposter-distroless",
	}

	engines := []engine.EngineType{
		"docker",
		"docker-all",
		"docker-distroless",
	}

	for _, engineType := range engines {
		actual := getImageRepo(engineType)
		if actual != expected[string(engineType)] {
			t.Errorf("Expected %s, got %s", expected[string(engineType)], actual)
		}
	}
}

func TestImages_getImageRepo_customReg(t *testing.T) {
	logger.SetLevel(logrus.TraceLevel)
	expectedPrefix := "test.repo/"
	expected := map[string]string{
		"docker":            expectedPrefix + "outofcoffee/imposter",
		"docker-all":        expectedPrefix + "outofcoffee/imposter-all",
		"docker-distroless": expectedPrefix + "outofcoffee/imposter-distroless",
	}

	engines := []engine.EngineType{
		"docker",
		"docker-all",
		"docker-distroless",
	}

	viper.Set("docker.registry", expectedPrefix)
	t.Cleanup(func() {
		viper.Set("docker.registry", nil)
	})

	for _, engineType := range engines {
		actual := getImageRepo(engineType)
		if actual != expected[string(engineType)] {
			t.Errorf("Expected %s, got %s", expected[string(engineType)], actual)
		}
	}
}
