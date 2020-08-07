GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test -v
GOCLEAN=$(GOCMD) clean
FOLDER_NAME= ogen-$(OGEN_VERSION)
OGEN_VERSION=0.0.1

build: 
	$(GOBUILD) 

build_cross_docker:
	DOCKER_BUILDKIT=1 docker build --file Dockerfile --output release .

build_cross: pack_linux_arm pack_linux_amd64 pack_linux_arm64 pack_osx_amd64 pack_windows_amd64

pack_linux_amd64: build_linux_amd64
	mkdir $(FOLDER_NAME)
	mv ogen ./$(FOLDER_NAME)
	tar -czvf ogen-$(OGEN_VERSION)-linux-amd64.tar.gz ./$(FOLDER_NAME)
	rm -r ./$(FOLDER_NAME)

build_linux_amd64:
	CC=x86_64-linux-gnu-gcc CXX=x86_64-linux-gnu-g++  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD)

pack_linux_arm64: build_linux_arm64
	mkdir $(FOLDER_NAME)
	mv ogen ./$(FOLDER_NAME)
	tar -czvf ogen-$(OGEN_VERSION)-linux-arm64.tar.gz ./$(FOLDER_NAME)
	rm -r ./$(FOLDER_NAME)

build_linux_arm64:
	CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ CGO_ENABLED=1 GOOS=linux GOARCH=arm64 $(GOBUILD)

pack_linux_arm: build_linux_arm
	mkdir $(FOLDER_NAME)
	mv ogen ./$(FOLDER_NAME)
	tar -czvf ogen-$(OGEN_VERSION)-linux-arm.tar.gz ./$(FOLDER_NAME)
	rm -r ./$(FOLDER_NAME)

build_linux_arm:
	CC=arm-linux-gnueabi-gcc CXX=arm-linux-gnueabi-g++ CGO_ENABLED=1 GOOS=linux GOARM=7 GOARCH=arm $(GOBUILD)

pack_osx_amd64: build_osx_amd64
	mkdir $(FOLDER_NAME)
	mv ogen ./$(FOLDER_NAME)
	tar -czvf ogen-$(OGEN_VERSION)-osx-amd64.tar.gz ./$(FOLDER_NAME)
	rm -r ./$(FOLDER_NAME)

build_osx_amd64:
	CXX=x86_64-apple-darwin19-clang++ CC=x86_64-apple-darwin19-clang CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GOBUILD)


pack_windows_amd64: build_windows_amd64
	mkdir $(FOLDER_NAME)
	mv ogen.exe ./$(FOLDER_NAME)
	zip -r ogen-$(OGEN_VERSION)-windows-amd64.zip ./$(FOLDER_NAME)
	rm -r ./$(FOLDER_NAME)

build_windows_amd64:
	CXX=x86_64-w64-mingw32-c++ CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GOBUILD) -tags netgo -ldflags '-extldflags "-static -static-libstdc++"' -a

gen_ssz:
	sszgen -path ./p2p/message.go
	sszgen -path ./p2p/msg_version.go
	sszgen -path ./p2p/msg_block.go -include ./primitives/block.go,./primitives/blockheader.go,./primitives/votes.go,./primitives/tx.go,./primitives/deposit.go,./primitives/exit.go,./primitives/slashing.go,./primitives/governance.go
	sszgen -path ./p2p/msg_addr.go
	sszgen -path ./p2p/msg_getblocks.go
	sszgen -path ./primitives/block.go -include ./bls/multisig.go,./primitives/votes.go,./primitives/blockheader.go,./primitives/tx.go,./primitives/deposit.go,./primitives/exit.go,./primitives/slashing.go,./primitives/governance.go
	sszgen -path ./primitives/blockheader.go
	sszgen -path ./primitives/coins.go
	sszgen -path ./primitives/deposit.go
	sszgen -path ./primitives/exit.go
	sszgen -path ./primitives/governance.go
	sszgen -path ./primitives/validator.go
	sszgen -path ./primitives/votes.go
	sszgen -path ./primitives/blockheader.go
	sszgen -path ./primitives/slashing.go -include ./primitives/votes.go,./primitives/blockheader.go
	sszgen -path ./primitives/tx.go -include ./bls/multisig.go
	sszgen -path ./primitives/state.go -objs SerializableState -include ./primitives/coins.go,./primitives/validator.go,./primitives/votes.go,./primitives/governance.go
	sszgen -path ./bls/combined.go
	sszgen -path ./bls/multisig.go
	sszgen -path ./chain/index/txs.go -objs TxLocator
	sszgen -path ./bdb/blocknodedisk.go
	sszgen -path ./peers/conflict/lastaction.go -objs ValidatorHelloMessage

clean:
	@echo Cleaning...
	$(GOCLEAN) ./...
	$(GOCLEAN) --testcache


