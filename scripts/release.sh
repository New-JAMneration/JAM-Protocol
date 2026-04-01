#!/bin/bash

GP_VERSION=${1:-"0.7.0"}
TARGET_VERSION=${2:-"0.1.0"}
OUTPUT=${3:-"new-jamneration-target"}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Use tar pipe to bypass BuildKit git-aware context issues.
# BuildKit (Docker 29+) uses git to filter the build context, which fails
# when gitignored files exist on disk or submodule .git pointer files are
# present. CI is unaffected (clean clone), but local builds need this.
tar -C "${PROJECT_ROOT}" \
  --exclude=.git \
  --exclude='*.exe' --exclude='*.dll' --exclude='*.so' --exclude='*.dylib' \
  --exclude=target --exclude=build --exclude=node_modules \
  --exclude='pkg/test_data' --exclude='*.log' \
  -c . | \
docker build \
  -f docker/Dockerfile \
  --build-arg GP_VERSION=${GP_VERSION} \
  --build-arg TARGET_VERSION=${TARGET_VERSION} \
  --build-arg OUTPUT=${OUTPUT} \
  -t new-jamneration-target -

docker create --name tempcontainer new-jamneration-target
docker cp tempcontainer:/build .
docker rm tempcontainer
