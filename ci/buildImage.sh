#!/bin/bash
IMAGE_NAME=oip

if [ -z $IMAGE_TAG ]; then
  IMAGE_TAG=dev
fi

mv build/linux/oipd bin/oipd
docker build -t "$IMAGE_NAME:$IMAGE_TAG" -f ci/Dockerfile.publish .