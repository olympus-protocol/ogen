FROM ubuntu:latest
ENV INSTALL_LLVM_VERSION=10.0.0-rc2
ENV DOCKER_CLI_EXPERIMENTAL=enabled
ENV GO=go1.14.4.linux-amd64.tar.gz
ENV DEBIAN_FRONTEND="noninteractive"
WORKDIR /build
RUN apt-get update && \
    apt-get install -y \
        curl xz-utils \
        gcc g++ mingw-w64 \
        gcc-aarch64-linux-gnu g++-aarch64-linux-gnu \
        cmake libssl-dev libxml2-dev vim apt-transport-https \
        zip unzip libtinfo5 patch zlib1g-dev autoconf libtool \
        pkg-config make docker.io gnupg2 libgmp-dev python

COPY clang_cross.sh ./
RUN ./clang_cross.sh
RUN ls ./
COPY osxcross.sh ./
RUN ./osxcross.sh


RUN git clone https://github.com/olympus-protocol/ogen && cd ogen && git fetch origin && git checkout bazel_builds
RUN cd /opt && curl https://storage.googleapis.com/golang/${GO} -o ${GO} && \
    tar zxf ${GO} && rm ${GO} && \
    ln -s /opt/go/bin/go /usr/bin/ && \
    export GOPATH=/root/go

## osxcross path and lib symlink
ENV PATH="/usr/x86_64-apple-darwin/osxcross/bin/:${PATH}"
RUN ln -s /usr/x86_64-apple-darwin/osxcross/lib/libtapi.so /usr/lib/libtapi.so
RUN ln -s /usr/x86_64-apple-darwin/osxcross/lib/libtapi.so.6 /usr/lib/libtapi.so.6
RUN ln -s /usr/x86_64-apple-darwin/osxcross/lib/libtapi.so.6.0.1 /usr/lib/libtapi.so.6.0.1    

RUN cd ogen && go mod download
## Working

## Temp
RUN cd ogen && git pull

## Linux amd64
RUN cd ogen && make clean && make pack_linux_amd64

## Linux arm64
RUN cd ogen && make clean && make pack_linux_arm64

## Darwin amd64
RUN cd ogen && make clean && make pack_osx_amd64

## Windows amd64
RUN cd ogen && make clean && make pack_windows_amd64

RUN ls ogen/ 
RUN ls ogen/*.tar.gz
RUN ls ogen/*.zip