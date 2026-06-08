#!/usr/bin/env sh
set -eu
if [ "$#" -lt 4 ]; then
  echo "usage: $0 <apk-file> <version> <version-code> <release-notes>" >&2
  exit 2
fi
apk="$1"; version="$2"; code="$3"; notes="$4"
./scripts/validate_apk.sh "$apk" >/dev/null
name=$(basename "$apk")
sha=$(./scripts/compute_checksum.sh "$apk")
size=$(wc -c < "$apk" | tr -d ' ')
mkdir -p api/v1 bin
cp "$apk" "bin/$name"
cat > api/v1/version.json <<JSON
{
  "version": "$version",
  "version_code": $code,
  "apk_filename": "$name",
  "apk_sha256": "$sha",
  "apk_size_bytes": $size,
  "release_notes": "$notes"
}
JSON
