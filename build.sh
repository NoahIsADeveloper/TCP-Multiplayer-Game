#!/usr/bin/env bash
set -euo pipefail

# Usage: ./build.sh windows|linux amd64|arm64|arm [optimize]
# optimize: "true" to enable -s -w, CGO_DISABLED and UPX packing. Anything else = no optimize.

if [ "${1:-}" = "" ] || [ "${2:-}" = "" ]; then
    echo "Usage: $0 windows|linux amd64|arm64|arm [optimize]"
    exit 1
fi

OS_ARG="$1"
ARCH_ARG="$2"
OPTIMIZE_ARG="${3:-false}"

case "$OS_ARG" in
    windows) GOOS="windows" ;;
    linux)   GOOS="linux" ;;
    *) echo "Invalid OS. Use 'windows' or 'linux'." ; exit 1 ;;
esac

case "$ARCH_ARG" in
    amd64)  GOARCH="amd64" ;;
    arm64)  GOARCH="arm64" ;;
    arm)    GOARCH="arm" ;;
    *) echo "Invalid arch. Use 'amd64', 'arm64', or 'arm'." ; exit 1 ;;
esac

GOARM="${GOARM:-7}"

BUILD_DIR="build"
mkdir -p "$BUILD_DIR"

OUTPUT="${BUILD_DIR}/build-${GOOS}-${GOARCH}"
if [ "$GOOS" = "windows" ]; then
    OUTPUT="${OUTPUT}.exe"
fi

echo "Building $OUTPUT for $GOOS-$GOARCH (optimize=${OPTIMIZE_ARG})..."

export GOOS
export GOARCH
export GOARM

if [ "$OPTIMIZE_ARG" = "true" ]; then
    export CGO_ENABLED=0
    LDFLAGS='-s -w'
    TRIMPATH='-trimpath'
else
    LDFLAGS=''
    TRIMPATH=''
fi

if [ -n "$LDFLAGS" ]; then
    go build -ldflags="$LDFLAGS" $TRIMPATH -o "$OUTPUT" ./src
else
    go build -o "$OUTPUT" ./src
fi

echo "go build finished: $(stat -c%s "$OUTPUT" 2>/dev/null || stat -f%z "$OUTPUT") bytes"

if [ "$OPTIMIZE_ARG" = "true" ]; then
    if command -v upx >/dev/null 2>&1; then
        echo "Compressing with upx (this may slow startup slightly)..."
        upx --best --ultra-brute "$OUTPUT" || {
            echo "UPX failed; leaving uncompressed binary."
        }
        echo "After UPX: $(stat -c%s "$OUTPUT" 2>/dev/null || stat -f%z "$OUTPUT") bytes"
    else
        echo "UPX not found. Install upx to pack the binary: https://upx.github.io/"
    fi
fi

echo "Build complete: $OUTPUT"
