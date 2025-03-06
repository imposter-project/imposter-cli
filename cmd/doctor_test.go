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

package cmd

import (
	"gatehill.io/imposter/internal/engine/docker"
	"gatehill.io/imposter/internal/engine/jvm"
	"github.com/stretchr/testify/require"
	"testing"
)

func init() {
	docker.EnableEngine()
	jvm.EnableSingleJarEngine()
	jvm.EnableUnpackedDistroEngine()
}

func Test_checkPrereqs(t *testing.T) {
	t.Run("check prereqs", func(t *testing.T) {
		got := checkPrereqs()
		require.Containsf(t, got, "You should be able to run Imposter", "summary should be present")
		require.Containsf(t, got, "Connected to Docker", "should find docker")
		require.Containsf(t, got, "Java version installed", "should find java")
	})
}
