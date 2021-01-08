#!/bin/bash

go get -u github.com/gobuffalo/packr/packr

packr

go mod tidy