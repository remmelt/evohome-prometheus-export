FROM golang:1.13.4 AS builder
LABEL maintainer="Remmelt Pit <remmelt@gmail.com>"

ENV GOPATH=/code/
RUN go get github.com/jcmturner/restclient github.com/stretchr/testify/assert

WORKDIR /code/src/github.com/remmelt/evohome-prometheus-export
ADD . .

RUN go test -v ./...
RUN go build -ldflags "-X main.buildstamp=`date -u '+%FT%T%Z'` -X main.githash=`git rev-parse HEAD`" \
    -tags netgo -o evohome-prometheus-export

FROM scratch
COPY --from=builder /code/src/github.com/remmelt/evohome-prometheus-export/evohome-prometheus-export /
COPY docker/security/DigiCertSHA2HighAssuranceServerCA.crt /DigiCertSHA2HighAssuranceServerCA.crt
ENV TRUST_CERT=/DigiCertSHA2HighAssuranceServerCA.crt
ENV SERVER_PORT=8080
ENTRYPOINT  [ "/evohome-prometheus-export" ]
