#!/bin/bash

go run github.com/gobuffalo/packr/v2/packr2
go build cmd/ogen/ogen.go

