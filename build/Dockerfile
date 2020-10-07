FROM ubuntu:latest as build

ENV INSTALL_LLVM_VERSION=10.0.0-rc2
ENV DOCKER_CLI_EXPERIMENTAL=enabled
ENV GO=go1.14.4.linux-amd64.tar.gz
ENV DEBIAN_FRONTEND="noninteractive"

WORKDIR /build

RUN apt-get update && \
    apt-get install -y \
        curl xz-utils \
        gcc g++ mingw-w64 \
        gcc-arm-linux-gnueabi g++-arm-linux-gnueabi \
        gcc-aarch64-linux-gnu g++-aarch64-linux-gnu \
        cmake libssl-dev libxml2-dev vim apt-transport-https \
        zip unzip libtinfo5 patch zlib1g-dev autoconf libtool \
        pkg-config make docker.io gnupg2 libgmp-dev python

RUN cd /opt && curl https://storage.googleapis.com/golang/${GO} -o ${GO} && \
    tar zxf ${GO} && rm ${GO} && \
    ln -s /opt/go/bin/go /usr/bin/ && \
    export GOPATH=/root/go

COPY ./build/clang_cross.sh ./
COPY ./build/osxcross.sh ./

RUN ./clang_cross.sh
RUN ./osxcross.sh

## osxcross path and lib symlink
ENV PATH="/usr/x86_64-apple-darwin/osxcross/bin/:${PATH}"
ENV PATH="/root/go/bin/:${PATH}"

RUN ln -s /usr/x86_64-apple-darwin/osxcross/lib/libtapi.so /usr/lib/libtapi.so
RUN ln -s /usr/x86_64-apple-darwin/osxcross/lib/libtapi.so.6 /usr/lib/libtapi.so.6
RUN ln -s /usr/x86_64-apple-darwin/osxcross/lib/libtapi.so.6.0.1 /usr/lib/libtapi.so.6.0.1

COPY .. /build/ogen
FROM eabz:ogen-builder as builder

RUN cd ogen && go mod download

## Build Cross-Compile
WORKDIR /build

RUN cd ogen && go get -u github.com/gobuffalo/packr/packr
RUN cd ogen && ./scripts/build-cross.sh