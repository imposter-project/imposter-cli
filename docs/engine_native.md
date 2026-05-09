# Using the native mock engine

Imposter supports different mock engine types: [Docker](./engine_docker.md), [JVM](./engine_jvm.md) and native. This document describes how to use the **native** engine.

The native engine is a lightweight, single-binary implementation of Imposter. It supports REST, OpenAPI, SOAP and gRPC mocking with JavaScript scripting.

> The native engine was previously named `golang`. The `golang` value is still accepted as a deprecated alias wherever the engine type can be configured.

## Prerequisites

No additional software is required. The native engine is downloaded automatically by the CLI.

## Features

The native engine supports:

- REST, OpenAPI (2.0 and 3.0+), SOAP (1.1 and 1.2) and gRPC plugins
- JavaScript scripting (Groovy is **not** supported)
- Request matching (path, query parameters, headers, body with JSONPath/XPath)
- Response templating with 30+ built-in functions
- Capture and stores (in-memory, Redis, DynamoDB)
- Fake data generation
- Rate limiting
- Performance simulation (delays)
- TLS and HTTP/2 support

## Configuration

### User default

The easiest way to set the engine type is to edit your user default [configuration](./config.md) in:

    $HOME/.imposter/config.yaml

Set the `engine` key to `native`:

```yaml
engine: native
```

### Environment variable

If you don't want to set your user defaults you can set the following environment variable:

    IMPOSTER_ENGINE=native

### Command line argument

You can also provide the `--engine-type` (or `-t`) command line argument to the `imposter up` command:

Example:

    imposter up --engine-type native

Or:

    imposter up -t native

## Differences from the JVM engine

- **No Groovy scripting** - use JavaScript instead
- **No passthrough/proxy responses** - cannot forward requests to upstream servers
- **No remote configuration sources** - configuration must be on the local filesystem
