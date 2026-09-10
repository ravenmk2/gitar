#!/usr/bin/env sh

mkdir -p bin

VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo dev)
SHA=$(git rev-parse --short HEAD 2>/dev/null || echo unknown)
LDFLAGS="-s -w -X gitar/pkg/app.Version=${VERSION} -X gitar/pkg/app.GitSHA=${SHA}"

GOOS=linux   GOARCH=amd64 go build -trimpath -ldflags "${LDFLAGS}" -o bin/gitar
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "${LDFLAGS}" -o bin/gitar.exe
