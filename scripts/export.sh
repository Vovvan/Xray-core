#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
repo=${1:?usage: export.sh <feature-repo> [base] [output-dir]}
base=${2:-v26.10.21}
out=${3:-"$root/patches"}

rm -rf "$out"
mkdir -p "$out"

git -C "$repo" format-patch -o "$out" "$base"..HEAD -- . ':(exclude).github'
(cd "$out" && ls -1 *.patch) > "$root/series"
