# Imposter SDK

The Imposter SDK allows you to embed Imposter directly into your own Go applications and workflows. Use it to programmatically start and manage mock servers, enabling powerful integration testing, local development environments, and custom tooling.

## Use cases

- **Integration testing** — spin up mock dependencies as part of your test suite
- **Local development** — run mock services alongside your application during development
- **Custom tooling** — build bespoke workflows that manage mock servers on demand
- **CI/CD pipelines** — programmatically control mocks in your build and deployment pipelines

## Key concepts

There are a few key concepts to learn before using the SDK:

- **configuration directory**: a directory containing a valid Imposter [configuration](https://docs.imposter.sh/configuration/)
- **engine type**: this can be `docker`, `jvm` or `golang` - see [Docker Engine](./engine_docker.md), [JVM Engine](./engine_jvm.md) or [Golang Engine](./engine_golang.md)
- **engine version**: this is the version of Imposter - see [Releases](https://github.com/imposter-project/imposter-jvm-engine/releases)

## Getting started

Import the Imposter SDK into your Go project:

```shell
go get github.com/imposter-project/imposter-cli
```

## Example

Here is a simple example that starts Imposter on port 8080, using the configuration in a given directory:

```go
package main

import (
    "sync"

    "github.com/imposter-project/imposter-cli/internal/engine"
    "github.com/imposter-project/imposter-cli/internal/engine/docker"
)

func main() {
    configDir := "/path/to/imposter/config"

    // can be docker, jvm or golang
    engineType := docker.EnableEngine()

    startOptions := engine.StartOptions{
        Port:           8080,
        Version:        "2.4.2",
        PullPolicy:     engine.PullIfNotPresent,
        LogLevel:       "DEBUG",
        ReplaceRunning: true,
    }

    mockEngine := engine.BuildEngine(engineType, configDir, startOptions)

    // block until the engine is terminated
    wg := &sync.WaitGroup{}
    mockEngine.Start(wg)
    wg.Wait()
}
```

## Learn more

- [Configuration reference](https://docs.imposter.sh/configuration/)
- [Docker Engine](./engine_docker.md)
- [JVM Engine](./engine_jvm.md)
- [Golang Engine](./engine_golang.md)
