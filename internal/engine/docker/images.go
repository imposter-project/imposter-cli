/*
Copyright © 2021 Pete Cornish <outofcoffee@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package docker

import (
	"context"
	"gatehill.io/imposter/internal/engine"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"os"
	"strings"
)

const defaultImageBase = "outofcoffee"

const dockerImageCore = "imposter"
const dockerImageAll = "imposter-all"
const dockerImageDistroless = "imposter-distroless"

type EngineImageProvider struct {
	engine.EngineMetadata
	imageAndTag string
}

func getProvider(engineType engine.EngineType, version string) *EngineImageProvider {
	return &EngineImageProvider{
		EngineMetadata: engine.EngineMetadata{
			EngineType: engineType,
			Version:    version,
		},
	}
}

func (d *EngineImageProvider) Provide(policy engine.PullPolicy) error {
	ctx, cli, err := buildCliClient()
	if err != nil {
		return err
	}
	imageAndTag, err := ensureContainerImage(cli, ctx, d.EngineType, d.Version, policy)
	if err != nil {
		return err
	}
	d.imageAndTag = imageAndTag
	return nil
}

func (d *EngineImageProvider) Satisfied() bool {
	return d.imageAndTag != ""
}

func (d *EngineImageProvider) GetEngineType() engine.EngineType {
	return d.EngineType
}

func ensureContainerImage(
	cli *client.Client,
	ctx context.Context,
	engineType engine.EngineType,
	imageTag string,
	imagePullPolicy engine.PullPolicy,
) (imageAndTag string, e error) {
	imageAndTag = getImageRepo(engineType) + ":" + imageTag

	if imagePullPolicy == engine.PullSkip {
		return imageAndTag, nil
	}

	if imagePullPolicy == engine.PullIfNotPresent {
		var hasImage = true
		_, _, err := cli.ImageInspectWithRaw(ctx, imageAndTag)
		if err != nil {
			if client.IsErrNotFound(err) {
				hasImage = false
			} else {
				return "", err
			}
		}
		if hasImage {
			logger.Debugf("engine image '%v' already present", imageTag)
			return imageAndTag, nil
		}
	}

	err := pullImage(cli, ctx, imageTag, imageAndTag)
	if err != nil {
		return "", err
	}
	return imageAndTag, nil
}

func pullImage(cli *client.Client, ctx context.Context, imageTag string, imageAndTag string) error {
	logger.Infof("pulling '%v' engine image", imageTag)
	reader, err := cli.ImagePull(ctx, imageAndTag, image.PullOptions{})
	if err != nil {
		return err
	}

	var pullLogDestination io.Writer
	if logger.IsLevelEnabled(logrus.TraceLevel) {
		pullLogDestination = os.Stdout
	} else {
		pullLogDestination = io.Discard
	}
	_, err = io.Copy(pullLogDestination, reader)
	if err != nil {
		return err
	}
	return nil
}

func getImageRepo(engineType engine.EngineType) string {
	var imageBase string
	registry := viper.GetString("docker.registry")
	if len(registry) > 0 {
		imageBase = registry
		if !strings.HasSuffix(imageBase, "/") {
			imageBase = imageBase + "/"
		}
		logger.Debugf("using docker registry: %s", registry)
	} else {
		imageBase = defaultImageBase + "/"
	}

	var imageName string
	switch engineType {
	case engine.EngineTypeDockerCore:
		imageName = dockerImageCore
		break
	case engine.EngineTypeDockerAll:
		imageName = dockerImageAll
		break
	case engine.EngineTypeDockerDistroless:
		imageName = dockerImageDistroless
		break
	default:
		panic("Unsupported engine type: " + engineType)
	}
	return imageBase + imageName
}
