#!/bin/bash

export DOCKER_BUILDKIT=1

mkdir -p release/
docker build --file build/Dockerfile --output release ./
