#!/bin/bash

echo "Downloading annotations"
mkdir -p api/proto_def/google/api
curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > api/proto_def/google/api/annotations.proto
curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > api/proto_def/google/api/http.proto

echo "Generating protocol buffer definitions"
protoc --proto_path=api/proto_def --go_out=plugins=grpc,paths=source_relative:api/proto api/proto_def/*.proto
protoc --proto_path=api/proto_def --grpc-gateway_out=logtostderr=true,paths=source_relative:api/proto api/proto_def/*.proto
protoc --proto_path=api/proto_def --swagger_out=allow_merge=true,merge_file_name=ogen,fqn_for_swagger_name=true,logtostderr=true:api/swagger api/proto_def/*.proto

rm -rf api/proto_def/google

