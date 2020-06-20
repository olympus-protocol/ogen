build_files:
	protoc -I. --go_out=plugins=grpc:./ --go_opt=paths=source_relative *.proto 
	ls ./*.pb.go | xargs -n1 -IX bash -c 'sed s/,omitempty// X > X.tmp && mv X{.tmp,}'
	ls ./*.pb.gw.go | xargs -n1 -IX bash -c 'sed s/,omitempty// X > X.tmp && mv X{.tmp,}'

build_proxy: 
	protoc -I. --grpc-gateway_out=logtostderr=true,paths=source_relative:./ *.proto

build: clean build_files build_proxy	

clean: 
	rm -rf ../chainrpc/proto/*.proto

build_doc:
	protoc --doc_out=./ --doc_opt=markdown,rpc.md *.proto
