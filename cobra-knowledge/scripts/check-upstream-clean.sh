#!/usr/bin/env sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)

for name in openclaw weknora openviking; do
  repo="$root/upstream/$name"
  if [ ! -d "$repo" ]; then
    echo "MISSING: $repo" >&2
    exit 1
  fi
  if git -C "$repo" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    if [ -n "$(git -C "$repo" status --porcelain)" ]; then
      echo "DIRTY: $repo" >&2
      git -C "$repo" status --short
      exit 1
    fi
    echo "CLEAN: $repo"
  else
    echo "SNAPSHOT: $repo has no git metadata; LeeClaw integration scripts must treat it as read-only."
  fi
done
