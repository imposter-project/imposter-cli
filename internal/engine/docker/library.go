package docker

import (
	"fmt"
	"gatehill.io/imposter/internal/engine"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"strings"
)

type DockerEngineLibrary struct {
	engineType engine.EngineType
}

func getLibrary(engineType engine.EngineType) *DockerEngineLibrary {
	return &DockerEngineLibrary{engineType}
}

func (DockerEngineLibrary) CheckPrereqs() (bool, []string) {
	var msgs []string
	ctx, cli, err := buildCliClient()
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("❌ Failed to build Docker client: %v", err))
		return false, msgs
	}

	version, err := cli.ServerVersion(ctx)
	if err != nil {
		if client.IsErrConnectionFailed(err) {
			msgs = append(msgs, fmt.Sprintf("❌ Failed to connect to Docker: %v", err))
			return false, msgs
		} else {
			msgs = append(msgs, fmt.Sprintf("❌ Failed to get Docker version: %v", err))
			return false, msgs
		}
	}
	msgs = append(msgs, "✅ Connected to Docker", fmt.Sprintf("✅ Docker version installed: %v", version.Version))

	return true, msgs
}

func (l DockerEngineLibrary) List() ([]engine.EngineMetadata, error) {
	ctx, cli, err := buildCliClient()
	if err != nil {
		return nil, fmt.Errorf("error building CLI client: %s", err)
	}
	imageRepo := getImageRepo(l.engineType)
	var available []engine.EngineMetadata
	imageSummaries, err := cli.ImageList(ctx, image.ListOptions{
		Filters: filters.NewArgs(filters.Arg("reference", imageRepo+":*")),
	})
	if err != nil {
		return nil, fmt.Errorf("error listing images: %s", err)
	}
	for _, imageSummary := range imageSummaries {
		for _, tag := range imageSummary.RepoTags {
			available = append(available, engine.EngineMetadata{
				EngineType: engine.EngineTypeDockerCore,
				Version:    strings.Split(tag, ":")[1],
			})
		}
	}
	return available, nil
}

func (l DockerEngineLibrary) GetProvider(version string) engine.Provider {
	return getProvider(l.engineType, version)
}

func (DockerEngineLibrary) IsSealedDistro() bool {
	return false
}

func (l DockerEngineLibrary) ShouldEnsurePlugins() bool {
	switch l.engineType {
	case engine.EngineTypeDockerCore:
		return true
	case engine.EngineTypeDockerAll:
		return false
	case engine.EngineTypeDockerDistroless:
		return true
	default:
		panic(fmt.Errorf("unsupported engine type: %s for Docker library", l.engineType))
	}
}
