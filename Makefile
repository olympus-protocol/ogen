GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
FOLDER_NAME= $(BINARY_NAME)-$(OGEN_VERSION)

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

update_deps:
	bazel run //:gazelle -- update-repos -from_file=go.mod -to_macro=deps.bzl%ogen_deps

build_cross: pack_linux_amd64 pack_linux_arm64 pack_osx_amd64 pack-windows-amd64

pack_linux_amd64: build_linux_amd64
	mkdir $(FOLDER_NAME)
	mv bazel_bin/_ogen/$(BINARY_NAME) ./$(FOLDER_NAME)
	tar -czvf ogen-$(OGEN_VERSION)-linux-amd64.tar.gz ./$(FOLDER_NAME)
	rm -r ./$(FOLDER_NAME)

build_linux_amd64:
	bazel-3.2.0 build //:ogen --config=linux_amd64_docker

pack_linux_arm64: build_linux_arm64
	mkdir $(FOLDER_NAME)
	mv bazel_bin/_ogen/$(BINARY_NAME) ./$(FOLDER_NAME)
	tar -czvf ogen-$(OGEN_VERSION)-linux-arm64.tar.gz ./$(FOLDER_NAME)
	rm -r ./$(FOLDER_NAME)

build_linux_arm64:
	bazel-3.2.0 build //:ogen --config=linux_arm64_docker

pack_osx_amd64: build_osx_amd64
	mkdir $(FOLDER_NAME)
	mv bazel_bin/_ogen/$(BINARY_NAME) ./$(FOLDER_NAME)
	tar -czvf ogen-$(OGEN_VERSION)-osx-amd64.tar.gz ./$(FOLDER_NAME)
	rm -r ./$(FOLDER_NAME)

build_osx_amd64:
	bazel-3.2.0 build //:ogen --config=osx_amd64_docker

pack-windows-amd64: build-windows-amd64
	mkdir $(FOLDER_NAME)
	mv bazel_bin/_ogen/$(BINARY_NAME) ./$(FOLDER_NAME)
	zip -r ogen-$(OGEN_VERSION)-windows-amd64.zip ./$(FOLDER_NAME)
	rm -r ./$(FOLDER_NAME)

build_windows_amd64:
	bazel-3.2.0 build //:ogen --config=windows_amd64_docker

clean:
	@echo Cleaning...
	$(GOCLEAN) ./
	rm -rf ./$(BINARY_NAME)
	rm -rf ./builds
	rm -rf ogen-darwin*
	rm -rf ogen-windows*
	rm -rf ogen-linux*
	rm -rf *.tar.gz
	rm -rf *.zip
	rm -rf chain.json
	rm -rf release/
	rm -rf bazel-*



