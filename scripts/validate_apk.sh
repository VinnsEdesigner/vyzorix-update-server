#!/usr/bin/env sh
set -eu
if [ "$#" -ne 1 ]; then
  echo "usage: $0 <apk-file>" >&2
  exit 2
fi
apk="$1"
[ -f "$apk" ] || { echo "missing APK: $apk" >&2; exit 1; }
case "$apk" in *.apk) ;; *) echo "file must end with .apk" >&2; exit 1;; esac
bytes=$(wc -c < "$apk" | tr -d ' ')
[ "$bytes" -gt 0 ] || { echo "APK is empty" >&2; exit 1; }
echo "ok $apk $bytes bytes $(sha256sum "$apk" | awk '{print $1}')"
