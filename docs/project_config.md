# Project configuration file

A project configuration file lives alongside your mocks and captures settings for that project — most commonly the engine version to pin, the plugins it needs, and environment variables to set. Because it sits in the project directory, you can commit it to version control so everyone working on the mocks gets the same behaviour.

When you run a command such as `imposter up` or `imposter bundle` in a directory, Imposter looks for a project configuration file there and merges it over your global [CLI configuration](./config.md).

## File name

The canonical file name is `imposter-project`, with a `.yaml`, `.yml` or `.json` extension:

    imposter-project.yaml

Running `imposter scaffold` creates an `imposter-project.yaml` file for you.

## Example

```yaml
# pin the engine version so everyone gets the same build
# see: https://github.com/imposter-project/imposter-jvm-engine/releases
version: "2.0.1"

# plugins to install for this project
plugins:
  - store-dynamodb

# environment variables to set when the mocks run
# see https://docs.imposter.sh/environment_variables/
env:
  IMPOSTER_LOG_LEVEL: DEBUG
```

A project configuration file accepts the same elements as the global CLI configuration file — see the [CLI configuration file](./config.md#cli-configuration-file) reference for the full set.

## How settings are merged

Settings are resolved in order of increasing precedence:

1. Your global CLI configuration file (`$HOME/.imposter/config.yaml`)
2. The project configuration file in the mock directory
3. Environment variables
4. Command line flags

So a value in the project file overrides your global default, and a command line flag overrides both.

## Backward compatibility

Earlier versions used a hidden `.imposter.yaml` (or `.yml`/`.json`) file. This name is still honoured, but is **deprecated**: when Imposter finds one it logs a warning prompting you to rename it to `imposter-project.yaml`.

If both a legacy `.imposter.*` and a canonical `imposter-project.*` file are present in the same directory, the canonical file takes precedence.
