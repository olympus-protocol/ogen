FROM eabz/ogen-builder:latest as build

COPY .. /build/ogen

RUN cd ogen && go mod download

## Build Cross-Compile
WORKDIR /build

RUN cd ogen && ./scripts/build-cross.sh

RUN mkdir /release && mv ogen/*.tar.gz /release && mv ogen/*.zip /release

FROM scratch as export
COPY --from=build /release/* .