#!/bin/bash

rm deployment.zip 2> /dev/null

echo "Building binary"
GOOS=linux GOARCH=amd64 go build -gccgoflags -static-libgo -o main main.go

echo "Create deployment package"
zip deployment.zip main

echo "Cleanup"
rm main
