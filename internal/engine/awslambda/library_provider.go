/*
Copyright © 2023 Pete Cornish <outofcoffee@gmail.com>

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

package awslambda

import (
	"gatehill.io/imposter/internal/engine"
	"gatehill.io/imposter/internal/logging"
)

type LambdaLibrary struct{}

type LambdaProvider struct {
	engine.EngineMetadata
}

var logger = logging.GetLogger()

var initialised = false

func EnableEngine() {
	if !initialised {
		initialised = true
		engine.RegisterLibrary(engine.EngineTypeAwsLambda, func() engine.EngineLibrary {
			return &LambdaLibrary{}
		})
	}
}

func (LambdaLibrary) GetProvider(version string) engine.Provider {
	return &LambdaProvider{
		EngineMetadata: engine.EngineMetadata{
			EngineType: engine.EngineTypeAwsLambda,
			Version:    version,
		},
	}
}

func (LambdaLibrary) IsSealedDistro() bool {
	return false
}

func (LambdaLibrary) ShouldEnsurePlugins() bool {
	return false
}

func (LambdaLibrary) CheckPrereqs() (bool, []string) {
	return true, []string{}
}

func (LambdaLibrary) List() ([]engine.EngineMetadata, error) {
	return []engine.EngineMetadata{}, nil
}

func (p *LambdaProvider) GetEngineType() engine.EngineType {
	return p.EngineType
}

func (*LambdaProvider) Provide(engine.PullPolicy) error {
	return nil
}

func (*LambdaProvider) Satisfied() bool {
	return true
}
