#!/usr/bin/env bash
# Cross-build the GMT fork for Home Assistant on a Raspberry Pi (linux/arm64)
# from Windows/macOS/Linux using Go and Zig as the C cross compiler.
#
# Needs: go >= 1.27, zig on PATH, and the Nabto SDK in .sdk/
#   .sdk/include/nabto/nabto_client.h (+ nabto_client_experimental.h)
#   .sdk/lib/libnabto_client.so        (linux-aarch64 build)
# from https://github.com/nabto/nabto-client-sdk-releases
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

VERSION="$(sed -n 's/^version: "\(.*\)"/\1/p' ha-addon-gmt/config.yaml)"
export CGO_ENABLED=1 GOOS=linux GOARCH=arm64
export CC="zig cc -target aarch64-linux-gnu.2.36"
export CXX="zig c++ -target aarch64-linux-gnu.2.36"

go build -ldflags="-s -w -X main.AppVersion=${VERSION}" -o ha-addon-gmt/steinel-bridge ./cmd/steinel-bridge
echo "built ha-addon-gmt/steinel-bridge (${VERSION})"
