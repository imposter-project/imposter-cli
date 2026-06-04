# Configuration

You can configure the Imposter CLI using command line arguments/flags or configuration files.

## Command line

Each command has its own list of arguments and flags, accessible using the `-h` flag, such as:

    imposter up -h

This prints the full set of flags, defaults, and usage for the command. The same applies to every subcommand (`imposter proxy -h`, `imposter scaffold -h`, and so on).

## Mock configuration files

Mocks are configured using files with the following suffixes:

* `-config.yaml`
* `-config.yml`
* `-config.json`

> For example: `orders-mock-config.yaml`

These files control behaviour such as responses, validation, scripting and more.

Learn about [Imposter mock configuration](https://docs.imposter.sh/configuration/) files.

## CLI Configuration file

You can also use a configuration file to set CLI defaults. By default, Imposter looks for a CLI configuration file located at `$HOME/.imposter/config.yaml`

> You can override the path to the CLI configuration file by passing the `--config CONFIG_PATH` flag.

The currently supported elements are as follows:

```yaml
# the engine type - valid values are "docker", "jvm" or "native"
# (the legacy value "golang" is still accepted as an alias for "native")
engine: "docker"

# the engine version - valid values are "latest", or a binary release such as "2.0.1"
# see: https://github.com/imposter-project/imposter-jvm-engine/releases
version: "latest"

# Docker engine specific configuration
docker:
  # bind mount flags
  # see: https://docs.docker.com/storage/bind-mounts
  bindFlags: ":z"

  # the container user (username or uid)
  containerUser: "imposter"

# JVM engine specific configuration
jvm:
  # override the path to the Imposter JAR file to use (default: automatically generated)
  jarFile: "/path/to/imposter.jar"
  
  # directory holding the JAR file cache (default: "$HOME/.imposter/cache")
  binCache: "/path/to/dir"

  # directory containing an unpacked Imposter distribution
  # note: this is generally only used by other tools
  distroDir: "/path/to/unpacked/distro"

# Plugin configuration
plugin:
  # override the directory holding plugin files
  dir: "/path/to/dir"

  # base directory holding versioned directories for plugin files (default: "$HOME/.imposter/plugins")
  # ignored if plugin.dir is set
  baseDir: "/path/to/base/dir"

# List of plugins to install
plugins:
  - store-dynamodb
  - store-redis

# Map of environment variables to set
env:
  IMPOSTER_EXAMPLE: "some-value"

cli:
  # the minimum required version of the CLI - not to be confused with engine version
  version: "0.40.0"
```

## Environment variables

Some configuration elements can be specified as environment variables:

- IMPOSTER_CLI_LOG_LEVEL
- IMPOSTER_ENGINE
- IMPOSTER_DOCKER_REGISTRY
- IMPOSTER_VERSION
- IMPOSTER_DEFAULT_PLUGINS
- IMPOSTER_DOCKER_BINDFLAGS
- IMPOSTER_DOCKER_CONTAINERUSER
- IMPOSTER_JVM_JARFILE
- IMPOSTER_JVM_BINCACHE
- IMPOSTER_JVM_DISTRODIR
- IMPOSTER_PLUGIN_BASEDIR
- IMPOSTER_PLUGIN_DIR

### Engine types

Imposter supports different mock engine types: Docker (default), JVM and native. For more information about configuring the engine type see:

- [Docker engine](./engine_docker.md) (default)
- [JVM engine](./engine_jvm.md)
- [Native engine](./engine_native.md)
