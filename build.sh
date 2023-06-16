#!/usr/bin/env bash

GIT_COMMIT=$(git rev-parse HEAD)
COMMIT_DATE=$(git log -1 --pretty=format:"%ci" | awk '{print $1}' | sed 's/-/./g')
VERSION="${VERSION:-${COMMIT_DATE}-${GIT_COMMIT}}"

os_list=("linux" "darwin" "windows")
arch_list=("arm64" "amd64")

function build() {
  echo "$(date +'%Y-%m-%d %H:%M:%S') - Building app..."
  CGO_ENABLED=0 go build -mod=vendor -ldflags "-X main.Version=${VERSION}" -o carctl cmd/*.go
  echo "$(date +'%Y-%m-%d %H:%M:%S') - App carctl build succeeded"
}

function build_all() {
  for os in "${os_list[@]}"; do
    for arch in "${arch_list[@]}"; do
      echo "$(date +'%Y-%m-%d %H:%M:%S') - Building for $os/$arch..."

      if [ "${os}" == 'windows' ]; then
        CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -mod=vendor -ldflags "-X main.Version=${VERSION}" -o "build/carctl/${os}/${arch}/carctl.exe" cmd/*.go
      else
        CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -mod=vendor -ldflags "-X main.Version=${VERSION}" -o "build/carctl/${os}/${arch}/carctl" cmd/*.go
      fi

      echo "$(date +'%Y-%m-%d %H:%M:%S') - $os/$arch build succeeded"
    done
  done
}

case "$1" in
all)
  build_all
  ;;
*)
  build
  ;;
esac
