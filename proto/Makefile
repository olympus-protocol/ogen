build:
	protoc --go_out=plugins=grpc:../chainrpc/proto --go_opt=paths=source_relative *.proto 

build_doc:
	protoc --doc_out=./ --doc_opt=markdown,rpc.md *.proto
