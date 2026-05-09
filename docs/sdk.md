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
- **engine type**: this can be `docker`, `jvm` or `native` - see [Docker Engine](./engine_docker.md), [JVM Engine](./engine_jvm.md) or [Native Engine](./engine_native.md)
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

    // register the engine implementation you want to use.
    // swap for jvm.EnableEngine() or native.EnableEngine() as required.
    docker.EnableEngine()

    startOptions := engine.StartOptions{
        Port:           8080,
        Version:        "latest",
        PullPolicy:     engine.PullIfNotPresent,
        LogLevel:       "DEBUG",
        ReplaceRunning: true,
    }

    mockEngine := engine.BuildEngine(engine.EngineTypeDockerCore, configDir, startOptions)

    // block until the engine is terminated
    wg := &sync.WaitGroup{}
    mockEngine.Start(wg)
    wg.Wait()
}
```

The matching engine type constants live on the `engine` package:

- `engine.EngineTypeDockerCore` (paired with `docker.EnableEngine()`)
- `engine.EngineTypeJvmSingleJar` (paired with `jvm.EnableEngine()`)
- `engine.EngineTypeNative` (paired with `native.EnableEngine()`)

## Learn more

- [Configuration reference](https://docs.imposter.sh/configuration/)
- [Docker Engine](./engine_docker.md)
- [JVM Engine](./engine_jvm.md)
- [Native Engine](./engine_native.md)
