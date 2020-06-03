build_proto_go:
	protoc --go_out=plugins=grpc:./ --go_opt=paths=source_relative *.proto 

build_proto_js:
	cd ogen-protojs && npm i
	protoc --plugin="protoc-gen-ts=./ogen-protojs/node_modules/.bin/protoc-gen-ts" --js_out="import_style=commonjs,binary:./ogen-protojs" --ts_out="service=grpc-web:./ogen-protojs" *.proto 

build_proto_py:
	protoc --python_out=./ogen-protopy ./*.proto