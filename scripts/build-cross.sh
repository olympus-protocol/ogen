#!/bin/bash

export OGEN_VERSION=0.2.0-alpha.2
export FOLDER_NAME=ogen-$OGEN_VERSION

go get -u github.com/gobuffalo/packr/packr

echo "Building linux_amd64"
GOOS=linux GOARCH=amd64 packr build cmd/ogen/ogen.go

mkdir $FOLDER_NAME
mv ogen ./$FOLDER_NAME

tar -czvf ogen-$OGEN_VERSION-linux-amd64.tar.gz ./$FOLDER_NAME
rm -r ./$FOLDER_NAME

echo "Building linux_arm64"
GOOS=linux GOARCH=arm64 packr build cmd/ogen/ogen.go

mkdir $FOLDER_NAME
mv ogen ./$FOLDER_NAME

tar -czvf ogen-$OGEN_VERSION-linux-arm64.tar.gz ./$FOLDER_NAME
rm -r ./$FOLDER_NAME

echo "Building darwin_amd64"
GOOS=darwin GOARCH=amd64 packr build cmd/ogen/ogen.go

mkdir $FOLDER_NAME
mv ogen ./$FOLDER_NAME

tar -czvf ogen-$OGEN_VERSION-osx-amd64.tar.gz ./$FOLDER_NAME
rm -r ./$FOLDER_NAME

echo "Building darwin_arm64"
GOOS=darwin GOARCH=arm64 packr build cmd/ogen/ogen.go

mkdir $FOLDER_NAME
mv ogen ./$FOLDER_NAME

tar -czvf ogen-$OGEN_VERSION-osx-arm64.tar.gz ./$FOLDER_NAME
rm -r ./$FOLDER_NAME

echo "Building windows_amd64"
GOOS=windows GOARCH=amd64 packr build cmd/ogen/ogen.go

mkdir $FOLDER_NAME
mv ogen.exe ./$FOLDER_NAME

zip -r ogen-$OGEN_VERSION-windows-amd64.zip ./$FOLDER_NAME
rm -r ./$FOLDER_NAME
