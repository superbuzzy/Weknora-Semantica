#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 3 ]; then
  echo "usage: $0 <openclaw-source> <weknora-source> <openviking-source>" >&2
  exit 2
fi
OPENCLAW="$1"
WEKNORA="$2"
OPENVIKING="$3"

need_file() { [ -f "$1" ] || { echo "FAIL missing: $1" >&2; exit 1; }; }
need_text() { grep -Fq "$2" "$1" || { echo "FAIL contract '$2' missing in $1" >&2; exit 1; }; }

need_file "$OPENCLAW/src/plugins/plugin-api.types.ts"
need_file "$OPENCLAW/src/plugins/tool-types.ts"
need_file "$OPENCLAW/src/plugins/hook-before-agent-start.types.ts"
need_file "$OPENCLAW/src/plugins/hook-types.ts"
need_file "$OPENCLAW/src/gateway/server-methods/client-types.ts"
need_file "$OPENCLAW/src/config/sessions/session-entry-provenance.ts"
need_text "$OPENCLAW/src/gateway/server-methods/client-types.ts" "authenticatedUserProfile?"
need_text "$OPENCLAW/src/gateway/server-methods/client-types.ts" "profileId: string"
need_text "$OPENCLAW/src/config/sessions/session-entry-provenance.ts" 'source: "profile" | "channel" | "unknown"'
need_text "$OPENCLAW/src/plugins/plugin-api.types.ts" "registerTool:"
need_text "$OPENCLAW/src/plugins/tool-types.ts" "sessionKey?: string"
need_text "$OPENCLAW/src/plugins/tool-types.ts" "sessionId?: string"
need_text "$OPENCLAW/src/plugins/hook-before-agent-start.types.ts" "toolsAllow?: string[]"
need_text "$OPENCLAW/src/plugins/hook-types.ts" "requiresToolAuthority?: true"
need_text "$OPENCLAW/src/plugins/hook-types.ts" "export type PluginHookToolAuthority"
need_text "$OPENCLAW/src/plugins/hook-types.ts" "allows(toolName: string): boolean"
need_text "$OPENCLAW/src/plugins/hook-types.ts" "assertActive(): void"
echo "PASS OpenClaw durable-profile/tool-authority contracts"

need_file "$WEKNORA/internal/middleware/auth.go"
need_file "$WEKNORA/internal/types/tenant_api_key.go"
need_text "$WEKNORA/internal/middleware/auth.go" "X-External-User-ID"
need_text "$WEKNORA/internal/middleware/auth.go" "RequireDirectHeader"
need_text "$WEKNORA/internal/types/tenant_api_key.go" "KnowledgeBaseIDs"
need_text "$WEKNORA/internal/types/tenant_api_key.go" "Capabilities"
if ! grep -R -Fq "knowledge-search" "$WEKNORA/internal"; then
  echo "FAIL WeKnora knowledge-search contract missing" >&2
  exit 1
fi
if ! grep -R -Fq "chunks/by-id" "$WEKNORA/internal"; then
  echo "FAIL WeKnora chunk evidence contract missing" >&2
  exit 1
fi
echo "PASS WeKnora external-principal/search/evidence contracts"

need_file "$OPENVIKING/openviking/server/auth/plugins/trusted.py"
need_file "$OPENVIKING/openviking/server/routers/skills.py"
need_file "$OPENVIKING/openviking/core/skill_loader.py"
need_file "$OPENVIKING/openviking/server/routers/sessions.py"
need_file "$OPENVIKING/openviking/server/routers/search.py"
need_text "$OPENVIKING/openviking/server/auth/plugins/trusted.py" "X-OpenViking-Account"
need_text "$OPENVIKING/openviking/server/auth/plugins/trusted.py" "X-OpenViking-User"
need_text "$OPENVIKING/openviking/server/routers/skills.py" "score_threshold"
need_text "$OPENVIKING/openviking/server/routers/skills.py" 'allowed_tools'
need_text "$OPENVIKING/openviking/core/skill_loader.py" 'allowed_tools_declared'
need_text "$OPENVIKING/openviking/core/skill_loader.py" '"allowed-tools"'
echo "PASS OpenViking trusted-principal/dynamic-skill contracts"

echo "PASS v0.8 upstream compatibility gate"
