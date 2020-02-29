#!/bin/sh
go run github.com/hellgate75/go-tcp-server/client -verbosity DEBUG transfer-file "./main.go" "~/tmpGoServer/main-remote.go"
