#!/bin/bash

go get -u github.com/gobuffalo/packr/packr
packr build cmd/ogen/ogen.go
