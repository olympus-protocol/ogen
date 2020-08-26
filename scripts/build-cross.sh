#!/bin/bash

export OGEN_VERSION=0.0.1
export FOLDER_NAME=ogen-$OGEN_VERSION

echo "Building linux_amd64"
GOOS=linux GOARCH=amd64 go build cmd/ogen/ogen.go
GOOS=linux GOARCH=amd64 go build cmd/ogen-cli/ogen-cli.go
GOOS=linux GOARCH=amd64 go build cmd/migration/migration.go

mkdir $FOLDER_NAME
mv ogen ./$FOLDER_NAME
mv ogen-cli ./$FOLDER_NAME
mv migration ./$FOLDER_NAME
tar -czvf ogen-$OGEN_VERSION-linux-amd64.tar.gz ./$FOLDER_NAME
rm -r ./$FOLDER_NAME

echo "Building linux_arm64"
GOOS=linux GOARCH=arm64 go build cmd/ogen/ogen.go
GOOS=linux GOARCH=arm64 go build cmd/ogen-cli/ogen-cli.go
GOOS=linux GOARCH=arm64 go build cmd/migration/migration.go

mkdir $FOLDER_NAME
mv ogen ./$FOLDER_NAME
mv ogen-cli ./$FOLDER_NAME
mv migration ./$FOLDER_NAME
tar -czvf ogen-$OGEN_VERSION-linux-arm64.tar.gz ./$FOLDER_NAME
rm -r ./$FOLDER_NAME

echo "Building linux_armv7"
GOOS=linux GOARM=7 GOARCH=arm go build cmd/ogen/ogen.go
GOOS=linux GOARM=7 GOARCH=arm go build cmd/ogen-cli/ogen-cli.go
GOOS=linux GOARM=7 GOARCH=arm go build cmd/migration/migration.go

mkdir $FOLDER_NAME
mv ogen ./$FOLDER_NAME
mv ogen-cli ./$FOLDER_NAME
mv migration ./$FOLDER_NAME
tar -czvf ogen-$OGEN_VERSION-linux-armv7.tar.gz ./$FOLDER_NAME
rm -r ./$FOLDER_NAME

echo "Building linux_armv6"
GOOS=linux GOARM=6 GOARCH=arm go build cmd/ogen/ogen.go
GOOS=linux GOARM=6 GOARCH=arm go build cmd/ogen-cli/ogen-cli.go
GOOS=linux GOARM=6 GOARCH=arm go build cmd/migration/migration.go

mkdir $FOLDER_NAME
mv ogen ./$FOLDER_NAME
mv ogen-cli ./$FOLDER_NAME
mv migration ./$FOLDER_NAME
tar -czvf ogen-$OGEN_VERSION-linux-armv6.tar.gz ./$FOLDER_NAME
rm -r ./$FOLDER_NAME

echo "Building darwin_amd64"
GOOS=darwin GOARCH=amd64 go build cmd/ogen/ogen.go
GOOS=darwin GOARCH=amd64 go build cmd/ogen-cli/ogen-cli.go
GOOS=darwin GOARCH=amd64 go build cmd/migration/migration.go

mkdir $FOLDER_NAME
mv ogen ./$FOLDER_NAME
mv ogen-cli ./$FOLDER_NAME
mv migration ./$FOLDER_NAME
tar -czvf ogen-$OGEN_VERSION-osx-amd64.tar.gz ./$FOLDER_NAME
rm -r ./$FOLDER_NAME

echo "Building windows_amd64"
GOOS=windows GOARCH=amd64 go build cmd/ogen/ogen.go
GOOS=windows GOARCH=amd64 go build cmd/ogen-cli/ogen-cli.go
GOOS=windows GOARCH=amd64 go build cmd/migration/migration.go

mkdir $FOLDER_NAME
mv ogen.exe ./$FOLDER_NAME
mv ogen-cli.exe ./$FOLDER_NAME
mv migration.exe ./$FOLDER_NAME
zip -r ogen-$OGEN_VERSION-windows-amd64.zip ./$FOLDER_NAME
rm -r ./$FOLDER_NAME