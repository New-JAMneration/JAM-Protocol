# You can build and run the Docker image with the following commands:
# docker build -t run-release-new-jamneration-target -f docker/run-release.Dockerfile .
# docker run --rm -u $(id -u):$(id -g) -v ./build/new-jamneration-target:/new-jamneration-target -v /tmp:/tmp run-new-jamneration-target /tmp/jam_target.sock
FROM alpine:latest
WORKDIR /

ENV USE_MINI_REDIS=true

ENTRYPOINT ["/new-jamneration-target"]

