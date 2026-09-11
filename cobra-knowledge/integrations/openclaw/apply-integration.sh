#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "usage: $0 <clean-openclaw-source> <derived-output>" >&2
  exit 2
fi

SOURCE="$(cd "$1" && pwd)"
OUTPUT="$2"
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"

if [ "$SOURCE" = "$(cd "$ROOT" 2>/dev/null && pwd || true)" ]; then
  echo "refusing to use the LeeClaw source tree as OpenClaw upstream" >&2
  exit 1
fi

for required in \
  package.json \
  pnpm-workspace.yaml \
  src/plugin-sdk/control-ui.ts \
  src/plugin-sdk/plugin-entry.ts; do
  if [ ! -f "$SOURCE/$required" ]; then
    echo "OpenClaw compatibility check failed: missing $required" >&2
    exit 1
  fi
done

if ! grep -Fq 'extensions/*' "$SOURCE/pnpm-workspace.yaml"; then
  echo "OpenClaw workspace no longer includes extensions/*" >&2
  exit 1
fi

rm -rf "$OUTPUT"
mkdir -p "$OUTPUT"
(
  cd "$SOURCE"
  tar --exclude='./.git' --exclude='./node_modules' -cf - .
) | tar -xf - -C "$OUTPUT"

copy_plugin() {
  local name="$1"
  local source="$ROOT/integrations/openclaw/$name-plugin"
  local target="$OUTPUT/extensions/leeclaw-$name"
  mkdir -p "$target"
  (
    cd "$source"
    tar --exclude='./node_modules' --exclude='./dist' -cf - .
  ) | tar -xf - -C "$target"
}

copy_plugin knowledge
copy_plugin openviking

cat > "$OUTPUT/LEECLAW-V0.7-DERIVED.txt" <<'TXT'
This is a derived OpenClaw build tree assembled by LeeClaw v0.7.
No upstream OpenClaw file was modified in place.
Added workspace extensions:
- extensions/leeclaw-knowledge
- extensions/leeclaw-openviking
OpenViking memory is connected by the identity-aware leeclaw-openviking hooks; do not configure the upstream static-user context-engine plugin in multi-user LeeClaw.
TXT

echo "PASS derived OpenClaw tree: $OUTPUT"
echo "PASS upstream source untouched: $SOURCE"
