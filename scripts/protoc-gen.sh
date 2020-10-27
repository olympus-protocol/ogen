#!/bin/bash
export VERSION=0.1.0

echo "Downloading annotations"
mkdir -p api/proto_def/google/api
mkdir -p api/proto_def/protoc-gen-openapiv2/options/

curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > api/proto_def/google/api/annotations.proto
curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > api/proto_def/google/api/http.proto
curl https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/master/protoc-gen-openapiv2/options/annotations.proto > api/proto_def/protoc-gen-openapiv2/options/annotations.proto
curl https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/master/protoc-gen-openapiv2/options/openapiv2.proto > api/proto_def/protoc-gen-openapiv2/options/openapiv2.proto

echo "Generating protocol buffer definitions"

protoc --proto_path=api/proto_def \
  --go_out ./api/proto --go_opt paths=source_relative api/proto_def/*.proto \
  --go-grpc_out ./api/proto --go-grpc_opt paths=source_relative api/proto_def/*.proto

protoc --proto_path=api/proto_def \
  --grpc-gateway_out ./api/proto --grpc-gateway_opt paths=source_relative api/proto_def/*.proto

protoc --proto_path=api/proto_def \
  --openapiv2_out ./api/proto \
  --openapiv2_opt allow_merge=true  --openapiv2_opt merge_file_name=ogen \
  --openapiv2_opt logtostderr=true \
  api/proto_def/*.proto

rm -rf api/proto_def/google
rm -rf api/proto_def/protoc-gen-openapiv2
