#!/bin/bash
# This script will run the release inside a Docker container.

docker build \
  -t run-release-new-jamneration-target \
  -f docker/run-release.Dockerfile .
docker run \
  --rm \
  -u $(id -u):$(id -g) \
  -v ./build/new-jamneration-target:/new-jamneration-target \
  -v /tmp:/tmp \
  run-release-new-jamneration-target /tmp/jam_target.sock
