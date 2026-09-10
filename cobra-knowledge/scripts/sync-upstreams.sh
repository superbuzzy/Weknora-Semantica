#!/usr/bin/env sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
update_repo() {
  dir=$1
  url=$2
  if [ -d "$dir/.git" ]; then
    git -C "$dir" fetch --all --prune
    branch=$(git -C "$dir" symbolic-ref --short HEAD)
    git -C "$dir" pull --ff-only origin "$branch"
  else
    echo "$dir is a source snapshot, not a git clone."
    echo "For future pull-based updates, replace it with: git clone $url $dir"
  fi
}
update_repo "$root/upstream/weknora" "https://github.com/Tencent/WeKnora.git"
update_repo "$root/upstream/semantica" "https://github.com/semantica-agi/semantica.git"
