# Release and publish

## Overview

We have Github action to build the release binary automatically when we create a new tag.  
When you create a new tag which follows our tag format `v*.*.*.*`, the Github action will be triggered to build the release binary and upload it to [our release repository](https://github.com/New-JAMneration/new-jamneration-release).  
If you find the Github action fails or you want to build and publish the release binary manually, you can follow the steps below.

> :warning: If the token expired, please create a new token with `repo` scope and update the secret `TOKEN` in the repository settings.

## Build the release binary

Please verify the version you intend to release.  
The version is defined in the project root directory.  
You can update the release version in the following files:  

- VERSION_GP
- VERSION_TARGET

### Prerequisite

- Git
- Docker

### Build with Makefile

```bash
make release-target
```

After running the command, you will get the release binary and compressed file in the build directory.

```bash
❯ tree -L 1 ./build
./build
├── new-jamneration-target
├── new-jamneration-target-linux-amd64-0.7.0.tar.gz
└── new-jamneration-target-linux-amd64-0.7.0.zip
```

### Run the release binary

You can run the release binary with following command:

```bash
make run-release-target
```

### Test with minifuzz

You can test against a running fuzz target using [minifuzz](https://github.com/davxy/jam-conformance/tree/main/fuzz-proto/minifuzz).

Start the binary with the same socket path Minifuzz will use (`JAM_FUZZ_SOCK_PATH` must match `--target-sock`), for example locally:

```bash
make run-target   # listens on /tmp/jam_target.sock per Makefile env
```

```bash
python minifuzz/minifuzz.py -d examples/v1/faulty --target-sock /tmp/jam_target.sock
python minifuzz/minifuzz.py -d examples/v1/forks --target-sock /tmp/jam_target.sock
python minifuzz/minifuzz.py -d examples/v1/no_forks --target-sock /tmp/jam_target.sock
```

Docker: after `make fuzz-docker-build`, use `scripts/run_fuzz_target_docker.sh` (host defaults to `<repo>/.jam_fuzz_docker_run`); point Minifuzz at `<host-mount>/fuzz.sock` (often `./.jam_fuzz_docker_run/fuzz.sock`).

## Create a release tag

Create a new tag in the main repository (JAM-Protocol) for the new release.

The tag format is:

```bash
v<GP_VERSION>.<TARGET_PITCH_VERSION>
```

if the Graypaper version is `0.7.0` and the target pitch version is `1`, the tag will be:

```bash
v0.7.0.1
```

In these situations, you need to update the target pitch version:

- Bug fixes that do not change the Graypaper version.
- Update logics (target cli) that do not belong to Graypaper specification.

## Publish the release binary

- Go to our release repository [New-JAMneration/new-jamneration-release](https://github.com/New-JAMneration/new-jamneration-release)
- Create a new release with the new tag and upload the compressed file in the build directory.
  - Write the Graypaper version and target version in the release notes.
  - Don't forget to update the release notes with the changes in this release.

Currently, our compressed file naming convention is:

```bash
new-jamneration-target-<OS>-<ARCH>-<GP_VERSION>.<EXT>
```

We will need to update the `GP_VERSION` when we have a stable release for a new version of Graypaper.  
Otherwise, we will only update the `TARGET_VERSION` for minor changes.

## Submit the target to jam-conformance

The [issue](https://github.com/davxy/jam-conformance/issues/123) is for discussing the submission of a new release to jam-conformance.
If davxy identifies any issues in our release, we will need to address them and create a new release.  
When submitting a new Graypaper version, update the `targets.json` file in the [our jam-conformance repository](https://github.com/New-JAMneration/jam-conformance).  
Open a pull request and wait for davxy to merge it.

The fuzz target listens on Unix socket via **environment variables** (`JAM_FUZZ`, `JAM_FUZZ_SPEC`, `JAM_FUZZ_DATA_PATH`, `JAM_FUZZ_SOCK_PATH`; see fuzz-proto README in jam-conformance). The published **Docker image** sets these defaults in its `Dockerfile`, so **`cmd`/`args` are not required** in `targets.json` (upstream also documents image-only submissions this way).

For example, update `targets.json` as follows:

```json
"new-jamneration": {
  "image": "ghcr.io/new-jamneration/new-jamneration-target:latest",
  "gp_version": "0.7.2"
}
```

### Verify the new release in jam-conformance

```bash
# DIR: jam-conformance
cd scripts
# Pull the Docker image
python target.py get new_jamneration
# Run the new target to verify it works correctly
python target.py run --no-docker new_jamneration

# or run with docker
# However, you maybe need to set a correct CPU set for your docker environment
# python target.py run new_jamneration
```
