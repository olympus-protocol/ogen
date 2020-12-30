FROM golang:1.16beta1-buster

WORKDIR /build

# TODO fix open ports needed for connections (inbound & outbound)

RUN git clone https://github.com/olympus-protocol/ogen

WORKDIR /build/ogen

RUN go get -u github.com/gobuffalo/packr/packr

RUN go build cmd/ogen/ogen.go

EXPOSE 8080 8081 24126

ENTRYPOINT ["./scripts/full-node/run.sh"]




