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
	"fmt"
	"gatehill.io/imposter/internal/engine"
	"gatehill.io/imposter/internal/stringutil"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	filters2 "github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

const labelKeyManaged = "io.gatehill.imposter.managed"
const labelKeyPort = "io.gatehill.imposter.port"
const labelKeyDir = "io.gatehill.imposter.dir"
const labelKeyHash = "io.gatehill.imposter.hash"

func genDefaultHash(absPath string, port int) string {
	return stringutil.Sha1hashString(fmt.Sprintf("%v:%d", absPath, port))
}

func findContainersWithLabels(cli *client.Client, ctx context.Context, labels map[string]string) ([]engine.ManagedMock, error) {
	filters := filters2.NewArgs()
	for key, value := range labels {
		filters.Add("label", fmt.Sprintf("%v=%v", key, value))
	}
	containers, err := cli.ContainerList(ctx, container.ListOptions{Filters: filters})
	if err != nil {
		return nil, err
	}
	logger.Tracef("containers matching labels: %v", containers)

	var mocks []engine.ManagedMock
	for _, container := range containers {
		mock := engine.ManagedMock{
			ID:   container.ID[0:12],
			Name: container.Names[0],
			Port: findPublicPort(container),
		}
		mocks = append(mocks, mock)
	}
	return mocks, nil
}

func findPublicPort(container types.Container) int {
	for _, port := range container.Ports {
		if port.PublicPort != 0 {
			return int(port.PublicPort)
		}
	}
	return 0
}
