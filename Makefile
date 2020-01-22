GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
OGEN_VERSION=0.1.0

ifeq ($(OS),Windows_NT)
    OS := Windows
else
    OS := $(shell uname)  # same as "uname -s"
endif

LOWECASE_OS = $(shell echo $(OS) | tr A-Z a-z)
FOLDER_NAME= $(BINARY_NAME)-$(LOWECASE_OS)-$(OGEN_VERSION)

ifeq ($(LOWECASE_OS),darwin)
    BINARYCLI_NAME := ogen-cli
    BINARY_NAME := ogen
else ifeq ($(LOWECASE_OS),linux)
    BINARYCLI_NAME := ogen-cli
    BINARY_NAME := ogen
else ifeq ($(LOWECASE_OS),windows)
    BINARYCLI_NAME := ogen-cli.exe
    BINARY_NAME := ogen.exe
endif

install-deps: clean
	@echo Install dependencies QT wallet for $(OS)
ifeq ($(LOWECASE_OS),darwin)
	./contrib/depends/install-osx.sh
else ifeq ($(LOWECASE_OS),linux)
	./contrib/depends/install-linux.sh
else ifeq ($(LOWECASE_OS),windows)
	./contrib/depends/install-windows.sh
else
	@echo No building specifications for $(OS)
endif

run: build
	@echo Running $(BINARY_NAME)
	./$(BINARY_NAME)

pack: build
	mkdir $(FOLDER_NAME)
	mv $(BINARY_NAME) ./$(FOLDER_NAME)
	mv $(BINARYCLI_NAME) ./$(FOLDER_NAME)
	tar -czvf ogen-$(LOWECASE_OS).tar.gz ./$(FOLDER_NAME)
	rm -r ./$(FOLDER_NAME)

build:
	@echo Building $(BINARY_NAME) for $(OS)
	$(GOBUILD) -o $(BINARY_NAME)
	$(GOBUILD) -o $(BINARYCLI_NAME) ./cli/.

clean:
	@echo Cleaning...
	$(GOCLEAN) ./
	$(GOCLEAN) ./cli/.
	rm -rf ./$(BINARY_NAME)
	rm -rf ./$(BINARYCLI_NAME)
	rm -rf ./builds
	rm -rf ogen-darwin*
	rm -rf ogen-windows*
	rm -rf ogen-linux*


