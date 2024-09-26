#!/bin/bash

platforms=(
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
  "darwin/amd64"
  "darwin/arm64"
)

for platform in "${platforms[@]}"; do
  OSARCH=(${platform//\// })
  OS=${OSARCH[0]}
  ARCH=${OSARCH[1]}
  echo "Building for $OS/$ARCH..."
  GOOS=$OS GOARCH=$ARCH go build -o "hm-$OS-$ARCH" # Output with platform in name
done

echo "Done!"
