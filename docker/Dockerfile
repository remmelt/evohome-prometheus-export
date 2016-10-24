FROM scratch
MAINTAINER Jonathan Turner <jt@jtnet.co.uk>

COPY build/output/evohome-prometheus-export  /evohome-prometheus-export
COPY security/DigiCertSHA2HighAssuranceServerCA.crt /DigiCertSHA2HighAssuranceServerCA.crt
ENV TRUST_CERT=/DigiCertSHA2HighAssuranceServerCA.crt

EXPOSE 8080
ENTRYPOINT  [ "/evohome-prometheus-export" ]
