#!/bin/bash

set -euo pipefail

if [ -z "${1:-}" ]; then
    echo "Usage: $0 windows|linux amd64|arm64|arm"
    exit 1
fi

if [ -z "${2:-}" ]; then
    echo "Usage: $0 windows|linux amd64|arm64|arm"
    exit 1
fi

OS_ARG="$1"
ARCH_ARG="$2"

case "$OS_ARG" in
    windows)
        GOOS="windows"
        ;;
    linux)
        GOOS="linux"
        ;;
    *)
        echo "Invalid OS. Use 'windows' or 'linux'."
        exit 1
        ;;
esac

case "$ARCH_ARG" in
    amd64|arm64|arm)
        GOARCH="$ARCH_ARG"
        ;;
    *)
        echo "Invalid architecture. Use 'amd64', 'arm64', or 'arm'."
        exit 1
        ;;
esac

OUTPUT="build/build-${GOOS}-${GOARCH}"
if [ "$GOOS" = "windows" ]; then
    OUTPUT="${OUTPUT}.exe"
fi

echo "Building $OUTPUT for $GOOS-$GOARCH..."

GOOS="$GOOS" GOARCH="$GOARCH" go build -o "$OUTPUT" ./src
if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Build complete: $OUTPUT"
