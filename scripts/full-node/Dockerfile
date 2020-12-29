FROM golang:1.16beta1-buster

WORKDIR /build

RUN git clone https://github.com/olympus-protocol/ogen

RUN cd ogen && ./scripts/build.sh
