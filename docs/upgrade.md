# Upgrade

Imposter can be [installed](./install.md) on Linux, macOS and Windows. This document explains how to upgrade Imposter to the latest version.

### Homebrew

If you installed Imposter using Homebrew, upgrade as follows:

    brew upgrade imposter

### Shell script

If you used the shell script approach (macOS and Linux only), you can re-run the script to upgrade:

```shell
curl -L https://raw.githubusercontent.com/imposter-project/imposter-cli/main/install/install_imposter.sh | bash -
```

> **Warning**
> It is good practice to examine [the script](../install/install_imposter.sh) first.

See [Releases](https://github.com/imposter-project/imposter-cli/releases) for the latest version.

## Manual upgrade

### macOS

Only ARM64 and Intel x86_64 are supported on macOS.

```shell
# see https://github.com/imposter-project/imposter-cli/releases
export IMPOSTER_CLI_VERSION=1.5.5

# choose one
export IMPOSTER_ARCH=arm64
#export IMPOSTER_ARCH=amd64

curl -L -o imposter-cli.tar.gz "https://github.com/imposter-project/imposter-cli/releases/download/v${IMPOSTER_CLI_VERSION}/imposter-cli_darwin_${IMPOSTER_ARCH}.tar.gz"
tar xvf imposter-cli.tar.gz
mv ./imposter /usr/local/bin/imposter
```

### Linux

Intel x86_64, ARM32 and ARM64 are supported on Linux.

```shell
# see https://github.com/imposter-project/imposter-cli/releases
export IMPOSTER_CLI_VERSION=1.5.5

# choose one
#export IMPOSTER_ARCH=arm64
#export IMPOSTER_ARCH=arm
export IMPOSTER_ARCH=amd64

curl -L -o imposter-cli.tar.gz "https://github.com/imposter-project/imposter-cli/releases/download/v${IMPOSTER_CLI_VERSION}/imposter-cli_linux_${IMPOSTER_ARCH}.tar.gz"
tar xvf imposter-cli.tar.gz
mv ./imposter /usr/local/bin/imposter
```

### Windows

Only Intel x86_64 is supported on Windows.

> These instructions assume `curl` and `unzip` are available. You can also download the ZIP archive from the [Releases](https://github.com/imposter-project/imposter-cli/releases) page.

```
# see https://github.com/imposter-project/imposter-cli/releases
SET IMPOSTER_CLI_VERSION=1.5.5

curl.exe --output imposter-cli.zip --url "https://github.com/imposter-project/imposter-cli/releases/download/v%IMPOSTER_CLI_VERSION%/imposter-cli_windows_amd64.zip"
unzip.exe imposter-cli.zip

# use command (or add to PATH)
imposter.exe [command/args]
```
