#!/usr/bin/env sh
set -eu
if [ "$#" -ne 1 ]; then
  echo "usage: $0 <apk-file>" >&2
  exit 2
fi
sha256sum "$1" | awk '{print $1}'
