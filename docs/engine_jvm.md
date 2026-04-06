# Using the JVM mock engine

Imposter supports different mock engine types: [Docker](./engine_docker.md), JVM and [Golang](./engine_golang.md). This document describes how to use the **JVM** engine.

## Prerequisites

Install a Java 11+ JVM.

For example, if you are using Homebrew install with:

    brew install openjdk@11

Or choose a distribution of your choice, such as [Eclipse Adoptium](https://adoptium.net/).

## Configuration

### User default

The easiest way to set the engine type is to edit your user default [configuration](./config.md) in:

    $HOME/.imposter/config.yaml

Set the `engine` key to `jvm`:

```yaml
engine: jvm
```

### Environment variable

If you don't want to set your user defaults you can set the following environment variable:

    IMPOSTER_ENGINE=jvm

### Command line argument

You can also provide the `--engine-type` (or `-t`) command line argument to the `imposter up` command:

Example:

    imposter up --engine-type jvm

Or:

    imposter up -t jvm
