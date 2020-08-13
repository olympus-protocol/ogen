#!/bin/bash

export DOCKER_BUILDKIT=1

docker build --file ../build/Dockerfile --output release ../
