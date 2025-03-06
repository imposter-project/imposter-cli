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
	"fmt"
	"gatehill.io/imposter/internal/debounce"
	"gatehill.io/imposter/internal/engine"
)

type DockerMockEngine struct {
	configDir   string
	options     engine.StartOptions
	provider    *EngineImageProvider
	containerId string
	debouncer   debounce.Debouncer
	shutDownC   chan bool
}

var initialised = false

func EnableEngine() {
	if !initialised {
		initialised = true
		register(engine.EngineTypeDockerCore)
		register(engine.EngineTypeDockerAll)
		register(engine.EngineTypeDockerDistroless)
	}
}

func register(engineType engine.EngineType) {
	engine.RegisterLibrary(engineType, func() engine.EngineLibrary {
		return getLibrary(engineType)
	})
	engine.RegisterEngine(engineType, func(configDir string, startOptions engine.StartOptions) engine.MockEngine {
		return buildEngine(engineType, configDir, startOptions)
	})
}

func buildEngine(engineType engine.EngineType, configDir string, options engine.StartOptions) engine.MockEngine {
	return &DockerMockEngine{
		configDir: configDir,
		options:   options,
		provider:  getProvider(engineType, options.Version),
		debouncer: debounce.Build(),
		shutDownC: make(chan bool),
	}
}

// Bundle implements the Docker engine steps to create a mock bundle.
// destFile is interpreted as the image tag.
func (d *EngineImageProvider) Bundle(configDir string, dest string) error {
	buf, err := addFilesToTar(configDir, d.imageAndTag)
	if err != nil {
		return fmt.Errorf("error adding files to build context: %v", err)
	}

	err = buildImage(buf, dest)
	if err != nil {
		return fmt.Errorf("error building image: %v", err)
	}

	return nil
}
