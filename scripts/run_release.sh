#!/bin/bash

docker run \
  --rm \
  -u $(id -u):$(id -g) \
  -v /tmp:/tmp \
  new-jamneration-target /tmp/jam_target.sock
