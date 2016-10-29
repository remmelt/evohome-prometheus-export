#!/bin/bash

export GOPATH=/tmp && \
cd /tmp && \
go get -d github.com/jcmturner/evohome-prometheus-export && \
go test -v src/github.com/jcmturner/evohome-prometheus-export/...
if [ $? -ne 0 ]; then
  echo "Golang tests failed"
  exit 1
fi
cd src/github.com/jcmturner/evohome-prometheus-export && \
go build -tags netgo && \
mv evohome-prometheus-export /tmp/output