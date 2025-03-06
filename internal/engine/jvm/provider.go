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

package jvm

import (
	"fmt"
	"gatehill.io/imposter/internal/debounce"
	"gatehill.io/imposter/internal/engine"
	"os/exec"
)

type JvmMockEngine struct {
	configDir string
	options   engine.StartOptions
	provider  *JvmProvider
	command   *exec.Cmd
	debouncer debounce.Debouncer
	shutDownC chan bool
}

type JvmProvider interface {
	engine.Provider
	GetStartCommand(args []string, env []string) *exec.Cmd
}

type JvmProviderOptions struct {
	engine.EngineMetadata
	javaCmd string
}

func buildEngine(configDir string, provider *JvmProvider, options engine.StartOptions) engine.MockEngine {
	return &JvmMockEngine{
		configDir: configDir,
		options:   options,
		provider:  provider,
		debouncer: debounce.Build(),
		shutDownC: make(chan bool),
	}
}

func (p *JvmProviderOptions) GetEngineType() engine.EngineType {
	return p.EngineType
}

func (p *JvmProviderOptions) Bundle(configDir string, dest string) error {
	return fmt.Errorf("JVM engine does not support bundling")
}
