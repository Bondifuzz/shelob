#!/usr/bin/env bash
set -euo pipefail

container_name="swaggerapi-petstore3"
image="swaggerapi/petstore3:1.0.7"
platform="linux/amd64"

if docker ps -a --format '{{.Names}}' | grep -qx "$container_name"; then
  docker rm -f "$container_name" >/dev/null
fi

docker run --platform "$platform" --ulimit nofile=1024:1024 --name "$container_name" -d -p 8080:8080 "$image"
