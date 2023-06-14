#!/usr/bin/env bash

GIT_COMMIT=$(git rev-parse HEAD)
COMMIT_DATE=$(git log -1 --pretty=format:"%ci" | awk '{print $1}' | sed 's/-/./g')
VERSION="${VERSION:-${COMMIT_DATE}-${GIT_COMMIT}}"

if uname -a | grep -qiE "(Microsoft|WSL)"; then
  echo "Build app on Windows, Version: ${VERSION}"
  go build -mod=vendor -ldflags "-X main.Version=${VERSION}" -o carctl.exe cmd/*.go
else
  echo "Build app on Unix, Version: ${VERSION}"
  go build -mod=vendor -ldflags "-X main.Version=${VERSION}" -o carctl cmd/*.go
  chmod +x carctl
fi
