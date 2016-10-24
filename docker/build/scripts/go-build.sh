#!/bin/bash

export GOPATH=/tmp && \
cd /tmp && \
go get -d github.com/jcmturner/evohome-prometheus-export && \
cd src/github.com/jcmturner/evohome-prometheus-export && \
go build -tags netgo && \
mv evohome-prometheus-export /tmp/output