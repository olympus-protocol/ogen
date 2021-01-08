#!/bin/bash

export OGEN_VERSION=0.2.0-alpha.3
export FOLDER_NAME=ogen-$OGEN_VERSION

xgo --branch bls  -pkg cmd/ogen --targets=windows-6.0/amd64,darwin-10.10/amd64,linux/arm64,linux/amd64 -out ogen-$OGEN_VERSION github.com/olympus-protocol/ogen

mkdir ./$FOLDER_NAME
mv ogen-$OGEN_VERSION-windows-6.0-amd64.exe ./$FOLDER_NAME/ogen.exe
zip -r ogen-$OGEN_VERSION-windows-amd64.zip ./$FOLDER_NAME
rm -rf ./$FOLDER_NAME

mkdir ./$FOLDER_NAME
mv ogen-$OGEN_VERSION-darwin-10.10-amd64 ./$FOLDER_NAME/ogen
tar -czvf ogen-$OGEN_VERSION-osx-amd64.tar.gz ./$FOLDER_NAME
rm -rf ./$FOLDER_NAME

mkdir ./$FOLDER_NAME
mv ogen-$OGEN_VERSION-linux-arm64 ./$FOLDER_NAME/ogen
tar -czvf ogen-$OGEN_VERSION-linux-arm64.tar.gz ./$FOLDER_NAME
rm -rf ./$FOLDER_NAME

mkdir ./$FOLDER_NAME
mv ogen-$OGEN_VERSION-linux-amd64 ./$FOLDER_NAME/ogen
tar -czvf ogen-$OGEN_VERSION-linux-amd64.tar.gz ./$FOLDER_NAME
rm -rf ./$FOLDER_NAME

gpg --detach-sign ogen-"$OGEN_VERSION"-linux-amd64.tar.gz
gpg --detach-sign ogen-"$OGEN_VERSION"-linux-arm64.tar.gz
gpg --detach-sign ogen-"$OGEN_VERSION"-osx-amd64.tar.gz
gpg --detach-sign ogen-"$OGEN_VERSION"-windows-amd64.zip
