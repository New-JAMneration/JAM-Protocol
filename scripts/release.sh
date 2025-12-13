#!/bin/bash

GP_VERSION=${1:-"0.7.0"}
TARGET_VERSION=${2:-"0.1.0"}
OUTPUT=${3:-"new-jamneration-target"}

docker build \
  --build-arg GP_VERSION=${GP_VERSION} \
  --build-arg TARGET_VERSION=${TARGET_VERSION} \
  --build-arg OUTPUT=${OUTPUT} \
  -t new-jamneration-target -f docker/Dockerfile .
docker create --name tempcontainer new-jamneration-target
docker cp tempcontainer:/build .
docker rm tempcontainer
