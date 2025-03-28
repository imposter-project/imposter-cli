# Configuration

You can configure the Imposter CLI using command line arguments/flags or configuration files.

## Command line

Each command has its own list of arguments and flags, accessible using the `-h` flag, such as:

```
$ imposter up -h
Starts a live mock of your APIs, using their Imposter configuration.

If CONFIG_DIR is not specified, the current working directory is used.

Usage:
  imposter up [CONFIG_DIR] [flags]

Flags:
      --auto-restart              Automatically restart when config dir contents change (default true)
      --deduplicate string        Override deduplication ID for replacement of containers
      --enable-file-cache         Enable file cache (default true)
      --enable-plugins            Enable plugins (default true)
  -t, --engine-type string        Imposter engine type (valid: docker,jvm - default "docker")
  -e, --env stringArray           Explicit environment variables to set
  -h, --help                      help for up
      --install-default-plugins   Install missing default plugins (default true)
      --mount-dir stringArray     (Docker engine type only) Extra directory bind-mounts in the form HOST_PATH:CONTAINER_PATH (e.g. $HOME/somedir:/opt/imposter/somedir) or simply HOST_PATH, which will mount the directory at /opt/imposter/<dir>
  -p, --port int                  Port on which to listen (default 8080)
      --pull                      Force engine pull
  -r, --recursive-config-scan     Scan for config files in subdirectories (default false)
  -s, --scaffold                  Scaffold Imposter configuration for all OpenAPI files
  -v, --version string            Imposter engine version (default "latest")
```

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
# the engine type - valid values are "docker" or "jvm"
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

# Default configuration regardless of engine version
default:
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

Imposter supports different mock engine types: Docker (default) and JVM. For more information about configuring the engine type see:

- [Docker engine](./docker_engine.md) (default)
- [JVM engine](./jvm_engine.md)
