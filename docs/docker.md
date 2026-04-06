# Docker

You can run the Imposter CLI as a Docker container, without installing it locally.

The container image is available at [`impostermocks/cli`](https://hub.docker.com/r/impostermocks/cli) on Docker Hub.

## Start mocks

To start a mock server using configuration in your current directory:

```shell
docker run --rm -v $PWD:/mocks -p 8080:8080 impostermocks/cli up
```

Your mock will be available at http://localhost:8080.

## Scaffold configuration

To generate mock configuration files in your current directory:

```shell
docker run --rm -v $PWD:/mocks impostermocks/cli scaffold
```

## Other commands

You can pass any CLI command and flags after the image name:

```shell
docker run --rm -v $PWD:/mocks impostermocks/cli --help
```
