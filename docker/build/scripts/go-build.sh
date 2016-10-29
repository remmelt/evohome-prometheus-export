#!/bin/bash

cd /tmp && \
git clone https://github.com/jcmturner/evohome-prometheus-export.git
cd evohome-prometheus-export && \
export GOPATH=$(pwd) && \
go test -v .
if [ $? -ne 0 ]; then
  echo "Golang tests failed"
  exit 1
fi
go build -tags netgo && \
mv evohome-prometheus-export /tmp/output