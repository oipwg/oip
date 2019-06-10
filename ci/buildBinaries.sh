#!/bin/bash
if [ -z $IMAGE_TAG ]; then
  IMAGE_TAG=dev
fi

mkdir build
mkdir build/linux
mkdir build/darwin
mkdir build/windows
docker build -t oip-build:$IMAGE_TAG -f ci/Dockerfile.build .
container_id=$(docker create oip-build:$IMAGE_TAG)
docker cp $container_id:/go/oipd.linux build/linux/oipd
docker cp $container_id:/go/oipd.darwin build/darwin/oipd
docker cp $container_id:/go/oipd.exe build/windows/oipd.exe
docker rm -v $container_id