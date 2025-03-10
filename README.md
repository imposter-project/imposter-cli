# Imposter: Scriptable, multipurpose mock server

Reliable, scriptable and extensible mock server for REST APIs, OpenAPI (and Swagger) specifications, SOAP/WSDL Web Services, Salesforce and HBase APIs.

> This project is the CLI tool for the [Imposter mock engine](https://www.imposter.sh).

Start a live mock of an OpenAPI specification with just:

```shell
$ imposter up -s

found 1 OpenAPI spec(s)
starting server on port 8080...
...
mock server up and running at http://localhost:8080
```

You now have a live mock of your OpenAPI spec running on localhost.

---

Or create a mock by proxying an exising endpoint:

```shell
$ imposter proxy https://example.com

starting proxy on port 8080
...
wrote response file /users.json for request GET /users
wrote config file example.com-config.yaml
```

Once you've recorded the HTTP exchanges, just run `imposter up` to start your mock. 

---

Or create a mock from an existing OpenAPI file:

```shell
$ imposter scaffold

found 1 OpenAPI spec(s)
generated 1 resources from spec
wrote Imposter config: /Users/mary/example/petstore-config.yaml
```

Just run `imposter up` to start your mock.

<img src="./docs/img/imposter-scaffold.gif" alt="Screenshot of scaffold command" width="67%">

---

## Features

- run standalone mocks in place of real systems
- turn an OpenAPI/Swagger file into a mock API for testing or QA (even before the real API is built)
- decouple your integration tests from the cloud/various back-end systems and take control of your dependencies
- validate your API requests against an OpenAPI specification
- capture data and validate later or use response templates to provide conditional responses
- proxy an existing endpoint to replay its responses as a mock

Send dynamic responses:

- Provide mock responses using static files or customise behaviour based on characteristics of the request.
- Power users can control mock responses with JavaScript or Java/Groovy script engines.
- Advanced users can write their own plugins in a JVM language of their choice.

## Getting started & documentation

You must have [Docker](https://docs.docker.com/get-docker/) installed and running, or if Docker is not available, you can run on the [JVM](./docs/jvm_engine.md).

### Installation

See the [Installation](./docs/install.md) instructions for your system or follow the quick-start instructions below:

#### Homebrew

If you have Homebrew installed:

    brew tap gatehill/imposter
    brew install imposter

#### Shell script

Or, use this one liner (macOS and Linux only):

```shell
curl -L https://raw.githubusercontent.com/imposter-project/imposter-cli/main/install/install_imposter.sh | bash -
```

## Usage

Top level command:

```
Usage:
  imposter [command]

Available Commands:
  up                Start live mocks of APIs
  scaffold          Create Imposter configuration from OpenAPI specs
  engine pull       Pull the engine into the cache
  engine list       List the engines in the cache
  doctor            Check prerequisites for running Imposter
  down              Stop running mocks
  list              List running mocks
  plugin install    Install plugin
  plugin list       List installed plugins
  proxy             Proxy an endpoint and record HTTP exchanges
  version           Print CLI version
  remote config     Configure remote
  remote deploy     Deploy active workspace
  remote show       Show remote
  remote status     Show remote status
  workspace delete  Delete a workspace
  workspace list    List all workspaces
  workspace new     Create a workspace
  workspace select  Set the active workspace
  help              Help about any command
```

### Create and start mocks

Example:

    imposter up

Usage:

```
Starts a live mock of your APIs, using their Imposter configuration.

If CONFIG_DIR is not specified, the current working directory is used.

Usage:
  imposter up [CONFIG_DIR] [flags]

Flags:
      --auto-restart              Automatically restart when config dir contents change (default true)
      --debug-mode                Enable JVM debug mode and listen on port 8000
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

### Generate Imposter configuration

Example:

    imposter scaffold

Usage:

```
Creates Imposter configuration files. If one or more OpenAPI/Swagger
specification files are present, they are used as the basis for the generated
resources. If no specification files are present, a simple REST mock is created.

If DIR is not specified, the current working directory is used.

Usage:
  imposter scaffold [DIR] [flags]

Flags:
  -f  --force-overwrite        Force overwrite of destination file(s) if already exist
      --generate-resources     Generate Imposter resources from OpenAPI paths (default true)
  -s  --script-engine string   Generate placeholder Imposter script (none|groovy|js) (default "none")
```

### Proxy HTTP(S) endpoint and record HTTP exchanges

Example:

    imposter proxy https://example.com

Usage:

```
Proxies an endpoint and records HTTP exchanges to file, in Imposter format.

Usage:
  imposter proxy [URL] [flags]

Flags:
      --capture-request-body        Capture the request body
      --capture-request-headers     Capture the request headers
      --flat                        Flatten the response file structure
  -h, --help                        help for proxy
  -i, --ignore-duplicate-requests   Ignore duplicate requests with same method and URI (default true)
  -o, --output-dir string           Directory in which HTTP exchanges are recorded (default: current working directory)
  -p, --port int                    Port on which to listen (default 8080)
  -H, --response-headers strings    Record only these response headers
  -r, --rewrite-urls                Rewrite upstream URL in response body to proxy URL
```

### Pull engine

Example:

    imposter engine pull

Usage:

```
Pulls a specified version of the engine binary/image into the cache.

If version is not specified, it defaults to 'latest'.

Usage:
  imposter engine pull [flags]

Flags:
  -t, --engine-type string    Imposter engine type (valid: docker,jvm - default "docker")
  -h, --help                  help for pull
  -f, --force                 Force engine pull
  -v, --version string        Imposter engine version (default "latest")
```

### List installed engines

Example:

    imposter engine list

Usage:

```
Lists all versions of engine binaries/images in the cache.

If engine type is not specified, it defaults to all.

Usage:
  imposter engine list [flags]

Flags:
  -t, --engine-type string   Imposter engine type (valid: docker,jvm - default is all
  -h, --help                 help for list
```

### Diagnose engine problems

```
Checks prerequisites for running Imposter, including those needed
by the engines.

Usage:
  imposter doctor
```

### Stop all running mocks

Example:

    imposter down

Usage:

```
Stops running Imposter mocks for the current engine type.

Usage:
  imposter down [flags]

Flags:
  -t, --engine-type string   Imposter engine type (valid: docker,jvm - default "docker")
  -h, --help                 help for down
```

### List all running mocks

Example:

    imposter list

Usage:

```
Lists running Imposter mocks for the current engine type
and reports their health.

Usage:
  imposter list [flags]

Aliases:
  list, ls

Flags:
  -t, --engine-type string   Imposter engine type (valid: docker,jvm - default "docker")
  -x, --exit-code-health     Set exit code based on mock health
  -h, --help                 help for list
  -q, --quiet                Quieten output; only print ID
```

#### Using as a healthcheck

You can use the `list` command as a healthcheck for running mocks.

```shell
$ imposter list --quiet --exit-code-health
```

This will return an exit code of `0` (success) if one or more mocks are running and healthy. If no mocks are running, or if one or more mock is unhealthy, a non-zero exit code will be returned.

> **Note**
> You can use the short versions of the arguments, so this can also be written:
> ```shell
> imposter list -qx
> ```

### Install plugin

Example:

    imposter plugin install [PLUGIN_NAME_1] [PLUGIN_NAME_N...]

Usage:

```
Installs plugins for a specific engine version.

If version is not specified, it defaults to 'latest'.

Example 1: Install named plugin

        imposter plugin install store-redis

Example 2: Install all plugins in config file

        imposter plugin install

Usage:
  imposter plugin install [PLUGIN_NAME_1] [PLUGIN_NAME_N...] [flags]

Flags:
  -h, --help             help for install
  -d, --save-default     Whether to save the plugin as a default
  -v, --version string   Imposter engine version (default "latest")
```

### List plugins

Example:

    imposter plugin list

Usage:

```
Lists all versions of installed plugins.

Usage:
  imposter plugin list [flags]

Aliases:
  list, ls

Flags:
  -v, --version string   Only show plugins for a specific engine version (default show all versions)
  -h, --help             help for list
```

### Help

```
Provides help for any command in the application.
Simply type imposter help [path to command] for full details.

Usage:
  imposter help [command] [flags]
```

## Logging

The default log level is `debug`. You can override this by setting the `LOG_LEVEL` environment variable:

    export LOG_LEVEL=info

...or passing the `--log-level <LEVEL>` argument, for example:

    imposter up --log-level trace

## Configuration

Learn more about [configuration](./docs/config.md).

---

## About Imposter

[Imposter](https://www.imposter.sh) is a mock server for REST APIs, OpenAPI (and Swagger) specifications, SOAP web services (and WSDL files), Salesforce and HBase APIs.

📖 **[Read the user documentation here](https://docs.imposter.sh)**

![Imposter logo](https://raw.githubusercontent.com/imposter-project/imposter-jvm-engine/main/docs/images/composite_logo13_cropped.png)

---

## Contributing

Suggestions and improvements to the CLI or documentation are welcome. Please raise pull requests targeting the `main` branch.
