# Docker Machine Driver of SMTX OS

[![Build Status](https://travis-ci.org/smartxworks/docker-machine-driver-smtxos.svg?branch=master)](https://travis-ci.org/smartxworks/docker-machine-driver-smtxos)

Create Docker machines remotely on [SMTX OS](https://www.smartx.com/smtx-os/). This driver requires SMTX OS 4.0 or above. Earlier versions of SMTX OS will not work with this driver.

## Installation

### From a release

The latest version of the driver binary is available on the [GithHub Releases](https://github.com/smartxworks/docker-machine-driver-smtxos/releases) page. Download the the binary that corresponds to your OS into a directory residing in your `$PATH`.

### From source

Make sure you have installed [Go](https://golang.org) and configured `$GOPATH` properly. For MacOS and Linux, make sure `$GOPATH/bin` is part of your `$PATH`. For Windows, make sure `%GOPATH%\bin` is included in `%PATH%`.

Run the following command:

```bash
go get -u github.com/smartxworks/docker-machine-driver-smtxos
```

## Usage

```bash
docker-machine create -d smtxos --smtxos-server <your-smtxos-server> --smtxos-password <your-smtxos-root-password> <machine-name>
```

## Options

```
docker-machine create -d smtxos --help
```

| CLI option                        | Environment variable          | Default value                     | Description |
| --------------------------------- | ----------------------------- | --------------------------------- | ----------- |
| `--smtxos-server`                 | `SMTXOS_SERVER`               | -                                 | address of SMTX OS server |
| `--smtxos-port`                   | `SMTXOS_PORT`                 | `80`                              | port of SMTX OS server |
| `--smtxos-username`               | `SMTXOS_USERNAME`             | `root`                            | username used to login SMTX OS |
| `--smtxos-password`               | `SMTXOS_PASSWORD`             | -                                 | password used to login SMTX OS |
| `--smtxos-cpu-count`              | `SMTXOS_CPU_COUNT`            | `2`                               | number of CPU cores for VM |
| `--smtxos-memory-size`            | `SMTXOS_MEMORY_SIZE`          | `4096`                            | size of memory for VM (in MB) |
| `--smtxos-disk-size`              | `SMTXOS_DISK_SIZE`            | `10240`                           | size of disk for VM (in MB) |
| `--smtxos-storage-policy-name`    | `SMTXOS_STORAGE_POLICY_NAME`  | `default`                         | name of storage policy of disk for VM |
| `--smtxos-dockeros-image-path`    | `SMTXOS_DOCKEROS_IMAGE_PATH`  | `[kubernetes]/SMTX-DockerOS.raw`  | path of DockerOS image on SMTX OS, in the format of `[datastore-name]/file-path` |
| `--smtxos-network-name`           | `SMTXOS_NETWORK_NAME`         | `default`                         | network name for VM |
| `--smtxos-ha`                     | `SMTXOS_HA`                   | `false`                           | whether to enable high availability for VM |

## License

[Apache 2.0](https://github.com/smartxworks/docker-machine-driver-smtxos/blob/master/LICENSE)
