#!/bin/bash

KIND_VERSION=v0.8.1
LINT_VERSION=v1.31.0


ACTUAL_KIND_VERSION=$(kind version | cut -d" " -f2)

if [[ $ACTUAL_KIND_VERSION =~ $KIND_VERSION ]]; then
    echo "kind is already installed"
else
  GO111MODULE="on" go get -u sigs.k8s.io/kind@${KIND_RELEASE}
fi

ACTUAL_LINT_VERSION=$(golangci-lint version 2>&1)
if [[ $ACTUAL_LINT_VERSION =~ $LINT_VERSION ]]; then
    echo "golanci-lint is already installed"
else
    echo "updating golanci-lint..."
    GO111MODULE=on CGO_ENABLED=0 go get github.com/golangci/golangci-lint/cmd/golangci-lint@${LINT_VERSION}
fi