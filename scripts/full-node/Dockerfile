FROM docker.pkg.github.com/olympus-protocol/ogen-builder/ogen-builder:main

WORKDIR /build

RUN git clone https://github.com/olympus-protocol/ogen

RUN cd ogen && ./scripts/build.sh