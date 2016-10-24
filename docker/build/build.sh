#!/bin/bash

BINDIR=$(cd `dirname $0` && pwd)

TAG=workingbuild-$$
echo "Building with tag: ${TAG}"
[ -d ${BINDIR}/output ] || mkdir  ${BINDIR}/output
rm -f ${BINDIR}/output/*

echo "Creating build docker image"
sudo docker build --no-cache=true -t ${TAG} --force-rm=true --rm=true ${BINDIR}
RETVAL=$?

if [ $RETVAL -ne 0 ]; then
  echo "Creating build docker image FAILED. Cleaning up..."
  sudo docker rmi ${TAG}
  sudo docker images | awk '/<none>/ {print $3}' | xargs sudo docker rmi
  echo "Creating build docker image FAILED"
  exit 1
else
  echo "Created build docker image: ${TAG}"
  echo "Running image to build the Go binary"
  sudo docker run -v ${BINDIR}/output:/tmp/output ${TAG}
  sudo docker ps -a | grep ${TAG} | awk '{print $1}' | xargs sudo docker rm
  sudo docker rmi ${TAG}
fi