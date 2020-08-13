#!/bin/bash

echo "Generating protocol buffer definitions"
#protoc -I. --proto_path=api/proto_def --go_out=plugins=grpc:./ --go_opt=paths=source_relative api/proto_def/*

#protoc --proto_path=api/proto_def --go_out=api/proto --go_opt=paths=source_relative api/proto_def/*
protoc -I/usr/local/include -I. -I"$GOPATH"/pkg/mod -I"$GOPATH"/pkg/mod/github.com/grpc-ecosystem/ --proto_path=api/proto_def --go_out=plugins=grpc,paths=source_relative:.api/proto api/proto_def/*

#protoc -I. --grpc-gateway_out=logtostderr=true,paths=source_relative:source_relative api/proto_def/*
