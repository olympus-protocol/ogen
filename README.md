# Ogen integration/staging tree

![Actions](https://github.com/olympus-protocol/ogen/workflows/Ogen/badge.svg)
[![Go Report](https://goreportcard.com/badge/github.com/olympus-protocol/ogen)](https://goreportcard.com/report/github.com/olympus-protocol/ogen)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/olympus-protocol/ogen?tab=doc)](https://pkg.go.dev/github.com/olympus-protocol/ogen?tab=doc)
[![codecov](https://codecov.io/gh/olympus-protocol/ogen/branch/master/graph/badge.svg)](https://codecov.io/gh/olympus-protocol/ogen)

> Ogen was a divine figure in classical antiquity to be the divine personification of the ocean.

Ogen is the main implementation of the Olympus protocol.

## Building

```bash
./scripts/build.sh
```

## Installing

```bash
bash <(wget --no-cache -qO- https://raw.githubusercontent.com/olympus-protocol/ogen/master/scripts/install.sh)
```

## Documentation

The complete documentation can be found here: <https://doc.oly.tech>

## Using Docker

### Run a full node
> Using this configuration can be used as a validator, but the keystore is not correctly stored on a persistent storage.

#### Pull the image
```
docker pull ghcr.io/olympus-protocol/ogen-full-node:latest
```
#### Run as a background service
```
docker run -p 80:8080 -p 81:8081 -d ghcr.io/olympus-protocol/ogen-full-node:latest
```

Now you will have a full-node instance running on the background with a dashboard exposed on port 80 and the REST API exposed on 81

### Run a full node with indexer
> Using this configuration will create a database storage inside the docker container, it is not persistent.

```
docker pull ghcr.io/olympus-protocol/ogen-indexer:latest
```

### Run a full node with validator
> Using this configuration will start a full node with a shared volume to store the keystore on a persistent storage.

```
docker pull ghcr.io/olympus-protocol/ogen-validator:latest
```