GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
FOLDER_NAME= $(BINARY_NAME)-$(OGEN_VERSION)
OGEN_VERSION=0.0.1

ifeq ($(OS),Windows_NT)
    OS := Windows
else
    OS := $(shell uname)  # same as "uname -s"
endif

LOWECASE_OS = $(shell echo $(OS) | tr A-Z a-z)

ifeq ($(LOWECASE_OS),darwin)
    BINARY_NAME := ogen
else ifeq ($(LOWECASE_OS),linux)
    BINARY_NAME := ogen
else ifeq ($(LOWECASE_OS),windows)
    BINARY_NAME := ogen.exe
endif

run: build
	@echo Running $(BINARY_NAME)
	./$(BINARY_NAME)

build: 
	@echo Building $(BINARY_NAME) for $(OS)
	$(GOBUILD) -o $(BINARY_NAME)

build_cross_docker:
	DOCKER_BUILDKIT=1 docker build ./

build_cross: pack_linux_amd64 pack_linux_arm64 pack_osx_amd64 pack_windows_amd64

pack_linux_amd64: build_linux_amd64
	mkdir $(FOLDER_NAME)
	mv $(BINARY_NAME) ./$(FOLDER_NAME)
	tar -czvf ogen-$(OGEN_VERSION)-linux-amd64.tar.gz ./$(FOLDER_NAME)
	rm -r ./$(FOLDER_NAME)

build_linux_amd64:
	CC=x86_64-linux-gnu-gcc CXX=x86_64-linux-gnu-g++  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)

pack_linux_arm64: build_linux_arm64
	mkdir $(FOLDER_NAME)
	mv $(BINARY_NAME) ./$(FOLDER_NAME)
	tar -czvf ogen-$(OGEN_VERSION)-linux-arm64.tar.gz ./$(FOLDER_NAME)
	rm -r ./$(FOLDER_NAME)

build_linux_arm64:
	CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ CGO_ENABLED=1 GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BINARY_NAME)

pack_osx_amd64: build_osx_amd64
	mkdir $(FOLDER_NAME)
	mv $(BINARY_NAME) ./$(FOLDER_NAME)
	tar -czvf ogen-$(OGEN_VERSION)-osx-amd64.tar.gz ./$(FOLDER_NAME)
	rm -r ./$(FOLDER_NAME)

build_osx_amd64:
	CXX=x86_64-apple-darwin19-clang++ CC=x86_64-apple-darwin19-clang CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)

pack_windows_amd64: build_windows_amd64
	mkdir $(FOLDER_NAME)
	mv $(BINARY_NAME) ./$(FOLDER_NAME)
	zip -r ogen-$(OGEN_VERSION)-windows-amd64.zip ./$(FOLDER_NAME)
	rm -r ./$(FOLDER_NAME)

build_windows_amd64:
	CXX=x86_64-w64-mingw32-c++ CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)

clean:
	@echo Cleaning...
	$(GOCLEAN) ./...
	rm -rf chain.json


