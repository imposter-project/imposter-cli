/*
Copyright © 2021 Pete Cornish <outofcoffee@gmail.com>

Licensed under the Apache License, Proxy 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"gatehill.io/imposter/internal/engine"
	"testing"
)

func Test_pull(t *testing.T) {
	type args struct {
		version    string
		engineType engine.EngineType
		pullPolicy engine.PullPolicy
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "pull latest docker always", args: args{
			version:    "latest",
			engineType: engine.EngineTypeDockerCore,
			pullPolicy: engine.PullAlways,
		}},
		{name: "pull specific docker version always", args: args{
			version:    "4.2.2",
			engineType: engine.EngineTypeDockerCore,
			pullPolicy: engine.PullAlways,
		}},
		{name: "pull specific docker version if not present", args: args{
			version:    "4.2.2",
			engineType: engine.EngineTypeDockerCore,
			pullPolicy: engine.PullIfNotPresent,
		}},
		{name: "pull docker all distro if not present", args: args{
			version:    "4.2.2",
			engineType: engine.EngineTypeDockerAll,
			pullPolicy: engine.PullIfNotPresent,
		}},
		{name: "pull docker all distro if not present", args: args{
			version:    "4.2.2",
			engineType: engine.EngineTypeDockerDistroless,
			pullPolicy: engine.PullIfNotPresent,
		}},
		{name: "pull latest jvm always", args: args{
			version:    "latest",
			engineType: engine.EngineTypeJvmSingleJar,
			pullPolicy: engine.PullAlways,
		}},
		{name: "pull specific jvm version always", args: args{
			version:    "4.2.2",
			engineType: engine.EngineTypeJvmSingleJar,
			pullPolicy: engine.PullAlways,
		}},
		{name: "pull specific jvm version if not present", args: args{
			version:    "4.2.2",
			engineType: engine.EngineTypeJvmSingleJar,
			pullPolicy: engine.PullIfNotPresent,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pull(tt.args.version, tt.args.engineType, tt.args.pullPolicy)
		})
	}
}
