#!/bin/bash

echo "Generating protocol buffer definitions"
protoc -I. --go_out=plugins=grpc:./api/proto --go_opt=paths=source_relative ./api/proto_def/*.proto
