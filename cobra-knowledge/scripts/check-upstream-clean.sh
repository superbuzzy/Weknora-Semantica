#!/usr/bin/env sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
for repo in "$root/upstream/weknora" "$root/upstream/semantica"; do
  if [ -d "$repo/.git" ]; then
    if [ -n "$(git -C "$repo" status --porcelain)" ]; then
      echo "DIRTY: $repo"
      git -C "$repo" status --short
      exit 1
    fi
    echo "CLEAN: $repo"
  else
    echo "SNAPSHOT: $repo has no .git metadata; CobraKnowledge does not write to it."
  fi
done
