FROM eabz/ogen-builder:latest as builder

RUN cd ogen && go mod download

WORKDIR /build

RUN cd ogen && go get -u github.com/gobuffalo/packr/packr
RUN cd ogen && ./scripts/build-cross.sh

RUN mkdir /release && mv ogen/*.tar.gz /release && mv ogen/*.zip /release

FROM scratch as export
COPY --from=build /release/* .