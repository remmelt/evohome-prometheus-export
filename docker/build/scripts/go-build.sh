#!/bin/bash

export GOPATH=/tmp && \
cd /tmp && \
go get github.com/jcmturner/evohome-prometheus-export && \
go get github.com/stretchr/testify/assert && \
cd src/github.com/jcmturner/evohome-prometheus-export && \
go test -v ./...
if [ $? -ne 0 ]; then
  echo "ERROR: Golang tests failed"
  exit 1
fi
go build -ldflags "-X main.buildstamp=`date -u '+%FT%T%Z'` -X main.githash=`git rev-parse HEAD`" -tags netgo
if [ $? -ne 0 ]; then
  echo "ERROR: Golang build failed"
  exit 1
fi
mv evohome-prometheus-export /tmp/output/
echo "Golang build completed successfully."
echo "Binary located in the output directory"