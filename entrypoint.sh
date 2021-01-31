#!/bin/sh
set -aeo pipefail

[ ! -f /etc/mqtt2prom.yaml ] && cp /go/src/mqtt2prom/mqtt2prom.yml /etc/mqtt2prom.yaml


if [ "$1" == "run" ]; then
  /go/bin/mqtt2prom --config /etc/mqtt2prom.yaml
else
  exec "$@"
fi
