# Using the Docker mock engine

Imposter supports different mock engine types: Docker and [JVM](./jvm_engine.md). This document describes how to use the **Docker** engine.

## Prerequisites

Install Docker: [https://docs.docker.com/get-docker/](https://docs.docker.com/get-docker/)

Ensure Docker is running. (You can check this with a `docker info`).

## Configuration

**Note: The Docker engine is the default, so you do not need to configure this explicitly.**

If you still want to specify which engine to use, follow these steps.

### User default

The easiest way to set the engine type is to edit your user default [configuration](./config.md) in:

    $HOME/.imposter/config.yaml

Set the `engine` key to `docker`:

```yaml
engine: docker
```

### Environment variable

If you don't want to set your user defaults you can set the following environment variable:

    IMPOSTER_ENGINE=docker

You can also change the registry to use by setting the `IMPOSTER_CLI_REGISTRY` environment variable:

    IMPOSTER_CLI_REGISTRY=ghcr.io/someorg/



### Command line argument

You can also provide the `--engine-type` (or `-t`) command line argument to the `imposter up` command:

Example:

    imposter up --engine-type docker

Or:

    imposter up -t docker
