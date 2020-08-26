FROM ubuntu:latest as build

ENV INSTALL_LLVM_VERSION=10.0.0-rc2
ENV DOCKER_CLI_EXPERIMENTAL=enabled
ENV GO=go1.14.4.linux-amd64.tar.gz
ENV DEBIAN_FRONTEND="noninteractive"

WORKDIR /build

RUN apt-get update && \
    apt-get install -y \
        curl xz-utils \
        cmake libssl-dev libxml2-dev vim apt-transport-https \
        zip unzip libtinfo5 patch zlib1g-dev autoconf libtool \
        pkg-config make docker.io gnupg2 libgmp-dev python

RUN cd /opt && curl https://storage.googleapis.com/golang/${GO} -o ${GO} && \
    tar zxf ${GO} && rm ${GO} && \
    ln -s /opt/go/bin/go /usr/bin/ && \
    export GOPATH=/root/go

COPY .. /build/ogen

RUN cd ogen && go mod download

## Build Cross-Compile

RUN cd ogen && ./scripts/build-cross.sh

RUN mkdir /release && mv ogen/*.tar.gz /release && mv ogen/*.zip /release

FROM scratch as export
COPY --from=build /release/* .