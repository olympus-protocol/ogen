FROM golang:1.16beta1-buster

WORKDIR /build

RUN git clone https://github.com/olympus-protocol/ogen

WORKDIR /build/ogen

RUN go get -u github.com/gobuffalo/packr/packr

RUN go build cmd/ogen/ogen.go

RUN ./ogen --dashboard

