#!/bin/bash

export DOCKER_BUILDKIT=1

cd build && docker build --file Dockerfile --output release .
mv release/ ../
