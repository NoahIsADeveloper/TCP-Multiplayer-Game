#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: $0 windows|linux|pi"
    exit 1
fi

case "$1" in
    windows)
        GOOS=windows
        GOARCH=amd64
        OUTPUT=build/build-windows.exe
        ;;
    linux)
        GOOS=linux
        GOARCH=arm64
        OUTPUT=build/build-linux
        ;;
    pi)
        GOOS=linux
        GOARCH=arm
        OUTPUT=build/build-pi
        ;;
    *)
        echo "Invalid argument. Use 'windows', 'linux', or 'pi'."
        exit 1
        ;;
esac

echo "Building for $GOOS..."

GOOS=$GOOS GOARCH=$GOARCH go build -o "$OUTPUT" ./src
if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit $?
fi

echo "Build complete: $OUTPUT"
