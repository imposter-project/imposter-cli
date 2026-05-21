# Imposter: Scriptable, multipurpose mock server

Reliable, scriptable and extensible mock server for REST APIs, OpenAPI (and Swagger) specifications, SOAP/WSDL Web Services, Salesforce and HBase APIs.

> This project is the CLI tool for the [Imposter mock engine](https://www.imposter.sh).

Stand up a live mock from an OpenAPI spec in one command:

```shell
$ imposter up -s

found 1 OpenAPI spec(s)
starting server on port 8080...
mock server up and running at http://localhost:8080
```

You now have a live mock of your OpenAPI spec running on localhost.

---

Record HTTP exchanges by proxying a real endpoint, then replay them as a mock:

```shell
$ imposter proxy https://example.com

starting proxy on port 8080
wrote response file /users.json for request GET /users
wrote config file example.com-config.yaml
```

Run `imposter up` to start the mock you just recorded.

---

Or scaffold a mock straight from an OpenAPI/WSDL file:

```shell
$ imposter scaffold

found 1 OpenAPI spec(s)
generated 1 resources from spec
wrote Imposter config: /Users/mary/example/petstore-config.yaml
```

Run `imposter up` to start your mock.

<img src="./docs/img/imposter-scaffold.gif" alt="Screenshot of scaffold command" width="67%">

---

## Features

- Stand up mocks in place of real systems, locally or in CI
- Turn an OpenAPI/Swagger or WSDL file into a working mock — even before the real API exists
- Decouple integration tests from cloud and back-end dependencies
- Validate requests against an OpenAPI specification
- Capture data and serve conditional, templated responses
- Store data for later retrieval or validation
- Proxy and record an existing endpoint to replay it as a mock
- Drive responses with JavaScript, Groovy, or your own plugin

## Install

You'll need [Docker](https://docs.docker.com/get-docker/), or alternatively a JVM ([JVM engine](./docs/engine_jvm.md)) or no extra runtime at all ([Native engine](./docs/engine_native.md)).

### Homebrew

```shell
brew tap imposter-project/imposter
brew install imposter
```

### Shell script (macOS and Linux)

```shell
curl -L https://raw.githubusercontent.com/imposter-project/imposter-cli/main/install/install_imposter.sh | bash -
```

For other platforms or manual installs, see [Installation](./docs/install.md).

## Common commands

Each command has full help via `imposter <command> --help`.

| Command | What it does |
| --- | --- |
| `imposter up [DIR]` | Start a live mock from Imposter config in `DIR` (defaults to current directory). Add `-s` to scaffold first. |
| `imposter scaffold [DIR]` | Generate Imposter config from any OpenAPI/Swagger or WSDL files in `DIR`. |
| `imposter proxy URL` | Forward traffic to `URL` and record each exchange to disk as a replayable mock. Add `--insecure` to skip TLS verification. |
| `imposter down ID` | Stop the mock with the given ID (see `imposter ls`). `-a` / `--all` stops every managed mock across all engine types. |
| `imposter list` | List running mocks and their health across all engine types. `-t` filters by engine type; `-qx` makes a tidy healthcheck. |
| `imposter bundle [DIR]` | Bundle config and engine into a Docker image or Lambda zip. |
| `imposter doctor` | Check that you have at least one engine ready to run. |
| `imposter engine pull` / `engine list` | Manage cached engine binaries and images. |
| `imposter plugin install` / `list` / `uninstall` | Manage engine plugins. |
| `imposter remote ...` | Configure and deploy to a remote Imposter target. |
| `imposter workspace ...` | Manage workspaces for remote deployments. |
| `imposter version` | Print CLI and engine version info. |

### Healthcheck

```shell
imposter list -qx
```

Exits `0` if at least one mock is running and healthy, non-zero otherwise.

## Logging

Default log level is `debug`. Override with the `LOG_LEVEL` environment variable:

```shell
export LOG_LEVEL=info
```

Or per invocation:

```shell
imposter up --log-level trace
```

## Configuration

See [Configuration](./docs/config.md) for the CLI config file, environment variables, and engine-specific settings.

Other deeper guides:

- [Docker engine](./docs/engine_docker.md) — the default
- [JVM engine](./docs/engine_jvm.md)
- [Native engine](./docs/engine_native.md)
- [Run the CLI itself in Docker](./docs/docker.md)
- [SDK — embed Imposter in your Go app](./docs/sdk.md)
- [Upgrade](./docs/upgrade.md)

## About Imposter

[Imposter](https://www.imposter.sh) is a mock server for REST APIs, OpenAPI/Swagger, SOAP/WSDL, Salesforce and HBase.

📖 **[Full user documentation](https://docs.imposter.sh)**

![Imposter logo](https://raw.githubusercontent.com/imposter-project/imposter-jvm-engine/main/docs/images/composite_logo13_cropped.png)

## Contributing

Suggestions and improvements to the CLI or its documentation are welcome. Please raise pull requests targeting the `main` branch.
