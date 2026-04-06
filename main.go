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

package main

import (
	"os"

	"github.com/imposter-project/imposter-cli/internal/config"
	"github.com/imposter-project/imposter-cli/internal/logging"
	"github.com/imposter-project/imposter-cli/internal/remote/awslambda"
	"github.com/imposter-project/imposter-cli/internal/remote/mockscloud"
	"github.com/imposter-project/imposter-cli/internal/stringutil"

	"github.com/imposter-project/imposter-cli/cmd"
	awslambdaengine "github.com/imposter-project/imposter-cli/internal/engine/awslambda"
	"github.com/imposter-project/imposter-cli/internal/engine/docker"
	"github.com/imposter-project/imposter-cli/internal/engine/golang"
	"github.com/imposter-project/imposter-cli/internal/engine/jvm"
)

const defaultLogLevel = "debug"

func main() {
	lvl := stringutil.GetFirstNonEmpty(os.Getenv("LOG_LEVEL"), os.Getenv("IMPOSTER_CLI_LOG_LEVEL"), defaultLogLevel)
	config.SetCliConfig(lvl)
	logging.SetLogLevel(lvl)

	// engines
	awslambdaengine.EnableEngine()
	docker.EnableEngine()
	jvm.EnableSingleJarEngine()
	jvm.EnableUnpackedDistroEngine()
	golang.EnableEngine()

	// remotes
	awslambda.Register()
	mockscloud.Register()

	cmd.Execute()
}
