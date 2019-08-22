#!/bin/bash
docker stop oip

docker rm oip
docker volume rm oip

./ci/buildBinaries.sh
./ci/buildImage.sh

docker volume create oip

docker run -d \
  --mount source=oip,target=/data \
  -p 1606:1606 -p 5601:5601 -p 9200:9200 \
  --env HTTP_USER=oip --env HTTP_PASSWORD=mypassword \
  --env NETWORK=testnet \
  --name=oip \
  oip:dev

docker logs --tail 10 -f oip