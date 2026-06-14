#!/bin/bash

DOCKER_PLATFORM="${DOCKER_PLATFORM:---platform linux/amd64}"

docker run \
  ${DOCKER_PLATFORM} \
  --rm \
  -u $(id -u):$(id -g) \
  -v /tmp:/tmp \
  -e JAM_FUZZ_SOCK_PATH=/tmp/jam_target.sock \
  new-jamneration-target
