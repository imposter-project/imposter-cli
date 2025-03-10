# Embed Imposter in your application

## Key concepts

There are a few key concepts to learn before starting Imposter:

- **configuration directory**: a directory containing a valid Imposter [configuration](https://docs.imposter.sh/configuration/)
- **engine type**: this can be `docker` or `jvm` - see [Docker Engine](./docker_engine.md) or [JVM Engine](./jvm_engine.md)
- **engine version**: this is the version of Imposter - see [Releases](https://github.com/imposter-project/imposter-jvm-engine/releases)

## Example

Here is a simple sample application that starts Imposter on port 8080, using the configuration in a given directory.

```go
package main

import "gatehill.io/imposter/internal/engine"
import "gatehill.io/imposter/internal/engine/docker"

func main() {
    configDir := "/path/to/imposter/config"

    // can be docker or jvm
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
