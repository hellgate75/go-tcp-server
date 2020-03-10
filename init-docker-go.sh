#!/bin/sh
export GO111MODULE=off
GITHUB_USER="hellgate75"
PROJECT_NAME="go-tcp-server"
BUILD_MODE="exe"
BUILD_TARGET="."
EXTENSION=""
BASE_FOLDER="$GOPATH/src/github.com/$GITHUB_USER"
PROJECT_FOLDER="$BASE_FOLDER/$PROJECT_NAME"
WORKDIR="$(pwd)"
echo "Working dir: $WORKDIR"
echo "Cpontent of dir: $WORKDIR"
ls -latr
echo "Creating base folder '$BASE_FOLDER' into folder: GOPATH '$GOPATH'"
mkdir -p $BASE_FOLDER
echo "Linking project folder: '$PROJECT_FOLDER' from source folder: '$WORKDIR'"
ln -s -T $WORKDIR $PROJECT_FOLDER
echo "Changing folder to $PROJECT_FOLDER"
cd $PROJECT_FOLDER
echo "Content of folder $PROJECT_FOLDER"
ls -latr
echo "Running go procedure into folder:$PROJECT_FOLDER"
go get -u github.com/golang/dep/...
dep init -v -skip-tools -no-examples
dep ensure -update -v
dep status ./... -f
#go mod init
#go mod tidy
echo "Testing project into folder:$PROJECT_FOLDER"
go test -v ./...
OUT_FILE_NAME="$PROJECT_NAME-$(uname -o|awk 'BEGIN {FS=OFS="/"}{print $NF}')-$(uname -m)$EXTENSION"
echo "Building project for making: $OUT_FILE_NAME"
go build -v -buildmode "$BUILD_MODE" -o "$OUT_FILE_NAME" $BUILD_TARGET
