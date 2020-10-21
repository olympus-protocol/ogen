#!/bin/bash

go run github.com/markbates/pkger/cmd/pkger -o cmd/ogen
go build cmd/ogen/ogen.go
