#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 2 ]; then echo "usage: $0 <clean-openclaw-source> <derived-output>" >&2; exit 2; fi
SOURCE="$(cd "$1" && pwd)"; OUTPUT="$2"; ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
if [ "$SOURCE" = "$(cd "$ROOT" 2>/dev/null && pwd || true)" ]; then echo "refusing to use the LeeClaw source tree as OpenClaw upstream" >&2; exit 1; fi
for required in package.json pnpm-workspace.yaml src/plugin-sdk/control-ui.ts src/plugin-sdk/plugin-entry.ts src/plugins/plugin-api.types.ts src/plugins/tool-types.ts src/plugins/hook-before-agent-start.types.ts src/plugins/hook-types.ts; do [ -f "$SOURCE/$required" ] || { echo "OpenClaw compatibility check failed: missing $required" >&2; exit 1; }; done
grep -Fq 'extensions/*' "$SOURCE/pnpm-workspace.yaml" || { echo "OpenClaw workspace no longer includes extensions/*" >&2; exit 1; }
grep -Fq 'packages/*' "$SOURCE/pnpm-workspace.yaml" || { echo "OpenClaw workspace no longer includes packages/*" >&2; exit 1; }
grep -Fq 'registerTool:' "$SOURCE/src/plugins/plugin-api.types.ts" || { echo "OpenClaw registerTool contract missing" >&2; exit 1; }
grep -Fq 'registerGatewayMethod' "$SOURCE/src/plugins/plugin-api.types.ts" || { echo "OpenClaw gateway method contract missing" >&2; exit 1; }
grep -Fq 'toolsAllow?: string[]' "$SOURCE/src/plugins/hook-before-agent-start.types.ts" || { echo "OpenClaw toolsAllow contract missing" >&2; exit 1; }
grep -Fq 'requiresToolAuthority?: true' "$SOURCE/src/plugins/hook-types.ts" || { echo "OpenClaw tool-authority contract missing" >&2; exit 1; }
rm -rf "$OUTPUT"; mkdir -p "$OUTPUT"; (cd "$SOURCE" && tar --exclude='./.git' --exclude='./node_modules' -cf - .) | tar -xf - -C "$OUTPUT"
copy_plugin(){ local name="$1"; local source="$ROOT/integrations/openclaw/$name-plugin"; local target="$OUTPUT/extensions/leeclaw-$name"; mkdir -p "$target"; (cd "$source" && tar --exclude='./node_modules' --exclude='./dist' -cf - .) | tar -xf - -C "$target"; }
copy_plugin knowledge
copy_plugin openviking
copy_plugin promotion
mkdir -p "$OUTPUT/packages/leeclaw-workspace-core"; (cd "$ROOT/integrations/openclaw/workspace-core" && tar --exclude='./node_modules' --exclude='./dist' -cf - .) | tar -xf - -C "$OUTPUT/packages/leeclaw-workspace-core"
cat > "$OUTPUT/LEECLAW-V0.10-DERIVED.txt" <<'TXT'
This is a derived OpenClaw build tree assembled by LeeClaw v0.10.
No upstream OpenClaw file was modified in place.

Added LeeClaw components:
- extensions/leeclaw-knowledge: read-only Agent Knowledge Runtime tools and management plane
- extensions/leeclaw-openviking: Memory / Experience / Dynamic Skill Runtime
- extensions/leeclaw-promotion: server-side reviewed Knowledge/Skill Promotion gateway
- packages/leeclaw-workspace-core: canonical Workspace / Role / Session scope resolver

OpenClaw remains the only human identity and final host tool authority.
WeKnora remains formal Knowledge Source of Truth.
OpenViking remains Memory / Experience / Skill Source of Truth.
LeeClaw Promotion stores candidate/review state only; Agent Tool Surface cannot publish formal assets.
TXT
echo "PASS derived OpenClaw v0.10 tree: $OUTPUT"
echo "PASS upstream source untouched: $SOURCE"
