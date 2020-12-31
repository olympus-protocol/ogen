FROM golang:1.16beta1-buster

WORKDIR /build

ENV DEBIAN_FRONTEND=noninteractive

RUN apt update && apt -y install postgresql

USER postgres

RUN service postgresql start && \
    psql --command "create database indexer;"

USER root

RUN git clone https://github.com/olympus-protocol/ogen

WORKDIR /build/ogen

RUN go get -u github.com/gobuffalo/packr/packr

RUN go build cmd/ogen/ogen.go

ENTRYPOINT ["./scripts/indexer/run.sh"]




