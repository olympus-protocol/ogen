FROM golang:1.16beta1-buster

WORKDIR /build

ENV DEBIAN_FRONTEND=noninteractive

RUN apt update && apt -y install postgresql

USER postgres

RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER indexer WITH SUPERUSER PASSWORD 'indexer';" &&\
    createdb -O indexer indexer

USER root

RUN git clone https://github.com/olympus-protocol/ogen

WORKDIR /build/ogen

RUN go build cmd/ogen/ogen.go

ENTRYPOINT ["./scripts/indexer/run.sh"]




