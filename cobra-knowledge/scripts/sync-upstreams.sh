#!/usr/bin/env sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)

if [ ! -f "$root/.gitmodules" ]; then
  echo "missing $root/.gitmodules; use this command from a full git clone" >&2
  exit 1
fi

git -C "$root" submodule sync --recursive
git -C "$root" submodule update --init --recursive

echo "Pinned upstream sources materialized:" 
git -C "$root" submodule status --recursive
