FROM ubuntu:18.04
ENV INSTALL_LLVM_VERSION=10.0.0-rc2
ENV DOCKER_CLI_EXPERIMENTAL=enabled

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
COPY osxcross.sh ./
RUN ./clang_cross.sh
RUN ./osxcross.sh
