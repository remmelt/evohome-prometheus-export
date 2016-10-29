#!/bin/bash

export GOPATH=/tmp && \
cd /tmp && \
go get github.com/jcmturner/evohome-prometheus-export && \
go get github.com/stretchr/testify/assert && \
cd src/github.com/jcmturner/evohome-prometheus-export && \
go test -v ./...
if [ $? -ne 0 ]; then
  echo "Golang tests failed"
  exit 1
fi
go build -tags netgo && \
mv evohome-prometheus-export /tmp/output