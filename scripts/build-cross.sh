#!/bin/bash

go run github.com/markbates/pkger/cmd/pkger -o cmd/ogen

export OGEN_VERSION=0.1.0-alpha.5
export FOLDER_NAME=ogen-$OGEN_VERSION

echo "Building linux_amd64"
CC=x86_64-linux-gnu-gcc CXX=x86_64-linux-gnu-g++  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build cmd/ogen/ogen.go

mkdir $FOLDER_NAME
mv ogen ./$FOLDER_NAME

tar -czvf ogen-$OGEN_VERSION-linux-amd64.tar.gz ./$FOLDER_NAME
rm -r ./$FOLDER_NAME

echo "Building linux_arm64"
CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build cmd/ogen/ogen.go

mkdir $FOLDER_NAME
mv ogen ./$FOLDER_NAME

tar -czvf ogen-$OGEN_VERSION-linux-arm64.tar.gz ./$FOLDER_NAME
rm -r ./$FOLDER_NAME

echo "Building darwin_amd64"
CXX=x86_64-apple-darwin19-clang++ CC=x86_64-apple-darwin19-clang CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build cmd/ogen/ogen.go

mkdir $FOLDER_NAME
mv ogen ./$FOLDER_NAME

tar -czvf ogen-$OGEN_VERSION-osx-amd64.tar.gz ./$FOLDER_NAME
rm -r ./$FOLDER_NAME

echo "Building windows_amd64"
CXX=x86_64-w64-mingw32-c++ CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -tags netgo -ldflags '-extldflags "-static -static-libstdc++"' -a cmd/ogen/ogen.go

mkdir $FOLDER_NAME
mv ogen.exe ./$FOLDER_NAME

zip -r ogen-$OGEN_VERSION-windows-amd64.zip ./$FOLDER_NAME
rm -r ./$FOLDER_NAME