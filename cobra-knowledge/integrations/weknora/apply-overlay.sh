#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "usage: $0 <weknora-upstream-dir> <derived-output-dir>" >&2
  exit 2
fi

src=$(cd "$1" && pwd)
dst=$2
root=$(cd "$(dirname "$0")" && pwd)

if [ ! -f "$src/frontend/src/views/knowledge/settings/GraphSettings.vue" ]; then
  echo "unsupported WeKnora source: GraphSettings.vue not found" >&2
  exit 1
fi

rm -rf "$dst"
mkdir -p "$dst"
# Build a derived tree. The upstream clone stays untouched and remains pullable.
tar --exclude='./.git' --exclude='./frontend/node_modules' -C "$src" -cf - . | tar -C "$dst" -xf -

while IFS= read -r -d '' file; do
  rel=${file#"$root/overlay/"}
  mkdir -p "$dst/$(dirname "$rel")"
  cp "$file" "$dst/$rel"
done < <(find "$root/overlay" -type f -print0)

for patch_file in "$root"/patches/*.patch; do
  [ -e "$patch_file" ] || continue
  patch --dry-run -s -p1 -d "$dst" < "$patch_file"
  patch -s -p1 -d "$dst" < "$patch_file"
done

echo "CobraKnowledge WeKnora overlay applied to: $dst"
