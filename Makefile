GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean

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



