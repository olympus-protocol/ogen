#!/bin/bash

export OGEN_VERSION=0.2.0-alpha.2

xgo --branch bls  -pkg cmd/ogen --targets=windows/amd64,darwin/amd64,linux/arm64,linux/amd64 -out ogen-$OGEN_VERSION github.com/olympus-protocol/ogen