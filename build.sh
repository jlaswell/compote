#!/usr/bin/env bash

set -euxo pipefail

CREATED=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
REVISION=$(git rev-parse HEAD)

docker build \
    --pull \
    --label "org.opencontainers.image.created=${CREATED}" \
	--label "org.opencontainers.image.revision=${REVISION}" \
    -t jlaswell/compote \
    .
