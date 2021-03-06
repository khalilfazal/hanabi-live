#!/bin/bash

set -e # Exit on any errors
set -x # Enable debugging

# This is the directory that this script lives in
DIRNAME=$(dirname "$0")

# Install VS Code extensions
code --install-extension ms-vscode.Go # For Golang
code --install-extension dbaeumer.vscode-eslint # For JavaScript

# Install the Golang dependencies for the project
cd $DIRNAME/../src
go get -u -v ./...

# Install the Golang development dependencies that VSCode uses
go get -u github.com/mdempsky/gocode
go get -u github.com/uudashr/gopkgs/cmd/gopkgs
go get -u github.com/lukehoban/go-outline
go get -u github.com/newhook/go-symbols
go get -u golang.org/x/tools/cmd/guru
go get -u golang.org/x/tools/cmd/gorename
go get -u github.com/fatih/gomodifytags
go get -u github.com/haya14busa/goplay/cmd/goplay
go get -u github.com/josharian/impl
go get -u github.com/tylerb/gotype-live
go get -u github.com/ianthehat/godef
go get -u golang.org/x/tools/cmd/godoc
go get -u github.com/zmb3/gogetdoc
go get -u golang.org/x/tools/cmd/goimports
go get -u github.com/sqs/goreturns
go get -u golang.org/x/lint/golint
go get -u github.com/cweill/gotests/...
go get -u github.com/alecthomas/gometalinter
go get -u honnef.co/go/tools/...
go get -u github.com/sourcegraph/go-langserver
go get -u github.com/derekparker/delve/cmd/dlv

# Install the Golang linter
go get -u github.com/golangci/golangci-lint || true # This is expected to give the "no Go files" error
cd $(go env GOPATH)/src/github.com/golangci/golangci-lint/cmd/golangci-lint
go install -ldflags "-X 'main.version=$(git describe --tags)' -X 'main.commit=$(git rev-parse --short HEAD)' -X 'main.date=$(date)'"

# Install the JavaScript linter
cd $DIRNAME/../public/js
npx install-peerdeps --dev eslint-config-airbnb-base
