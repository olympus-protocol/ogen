build_files:
	protoc -I. --go_out=plugins=grpc:../chainrpc/proto --go_opt=paths=source_relative *.proto 

build_proxy: 
	protoc -I. --grpc-gateway_out=logtostderr=true,paths=source_relative:../chainrpc/proto/ *.proto

build: clean build_files build_proxy	

clean: 
	rm -rf ../chainrpc/proto/*.proto

build_doc:
	protoc --doc_out=./ --doc_opt=markdown,rpc.md *.proto
