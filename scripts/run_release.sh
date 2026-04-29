#!/bin/bash

docker run \
  --rm \
  -u $(id -u):$(id -g) \
  -v /tmp:/tmp \
  -e JAM_FUZZ_SOCK_PATH=/tmp/jam_target.sock \
  new-jamneration-target
