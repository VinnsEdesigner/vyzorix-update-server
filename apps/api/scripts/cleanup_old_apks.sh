#!/usr/bin/env sh
set -eu
keep="${1:-3}"
dir="${2:-bin}"
[ -d "$dir" ] || exit 0
find "$dir" -maxdepth 1 -type f -name '*.apk' -printf '%T@ %p\n' \
  | sort -nr \
  | awk -v keep="$keep" 'NR > keep { $1=""; sub(/^ /, ""); print }' \
  | while IFS= read -r apk; do
      [ -n "$apk" ] && rm -f -- "$apk"
    done
