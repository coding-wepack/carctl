#!/usr/bin/env bash

VERSION="${VERSION:-${CODING_COMMIT}}"

go build -mod=vendor -ldflags "-X main.Version=${VERSION}" -o carctl cmd/*.go
