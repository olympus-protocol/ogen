
OGEN_VERSION=1.0.0

build:
	xgo --targets=windows/amd64,linux/arm64,linux/amd64,darwin/amd64 github.com/olympus-protocol/ogen

pack_release: build
	mkdir release
