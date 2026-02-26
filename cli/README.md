# ADB CLI

A command-line interface for managing and querying the Activist Database (ADB).

## Usage

```
adb [command]
```

Run `adb --help`, or any subcommand with `--help`, to see available commands.

If not using the devcontainer, be sure to set environment variables defined in `adb --help`. The devcontainer already
defines all the required variables.

Note that the devcontainer provides a bash alias to run `adb` without explicitly building any binary file.

## Docker image

When using the devcontainer, there isn't a need to run the docker image, but building and running the containerized CLI
locally may be useful to test the docker image before deploying it for production use.

### Building the docker image

```sh
docker build . -f Dockerfile.cli -t dxe/adb-cli
```

Or via the Makefile:

```sh
make prod_build
```

### Running dockerized CLI locally

When running this command from inside the devcontainer, add `--network host` so the CLI container shares the
devcontainer network namespace and can resolve the mysql database host.
The devcontainer defines the environment variables in the below command so there's no need to enter their values.
The image has no entrypoint arguments by default, so pass the command and any flags directly.

```sh
docker run --rm \
  --network host \
  -e DB_USER \
  -e DB_PASSWORD \
  -e DB_NAME \
  -e DB_PROTOCOL \
  dxe/adb-cli [command] [flags]
```
