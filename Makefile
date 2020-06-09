GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
OGEN_VERSION=1.0.0

ifeq ($(OS),Windows_NT)
    OS := Windows
else
    OS := $(shell uname)  # same as "uname -s"
endif

LOWECASE_OS = $(shell echo $(OS) | tr A-Z a-z)
FOLDER_NAME= $(BINARY_NAME)-$(OGEN_VERSION)

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

pack: build
	mkdir $(FOLDER_NAME)
	mv $(BINARY_NAME) ./$(FOLDER_NAME)
	tar -czvf ogen-$(OGEN_VERSION)-$(LOWECASE_OS).tar.gz ./$(FOLDER_NAME)
	rm -r ./$(FOLDER_NAME)	

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
	rm -rf release/
	rm -rf ogen-protopy*




