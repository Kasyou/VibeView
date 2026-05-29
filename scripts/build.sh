#!/bin/bash
set -e

APP="vibeview"
VERSION="${1:-0.1.0}"
BUILD_DIR="build"

rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

PLATFORMS=(
  "windows/amd64/.exe"
  "darwin/amd64/"
  "darwin/arm64/"
  "linux/amd64/"
  "linux/arm64/"
)

for entry in "${PLATFORMS[@]}"; do
  IFS='/' read -r GOOS GOARCH EXT <<< "$entry"
  OUTPUT="$BUILD_DIR/${APP}_${VERSION}_${GOOS}_${GOARCH}${EXT}"
  echo "Building $OUTPUT ..."
  GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "$OUTPUT" .
done

echo "Done. Builds in $BUILD_DIR/"
ls -la "$BUILD_DIR/"
