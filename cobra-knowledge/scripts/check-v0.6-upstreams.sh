#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 3 ]; then
  echo "usage: $0 <openclaw-source> <weknora-source> <openviking-source>" >&2
  exit 2
fi
OPENCLAW="$1"; WEKNORA="$2"; OPENVIKING="$3"
need_file(){ [ -f "$1" ] || { echo "FAIL missing: $1" >&2; exit 1; }; }
need_text(){ grep -Fq "$2" "$1" || { echo "FAIL contract '$2' missing in $1" >&2; exit 1; }; }

need_file "$OPENCLAW/src/plugins/plugin-api.types.ts"
need_file "$OPENCLAW/src/gateway/server-methods/client-types.ts"
need_file "$OPENCLAW/src/config/sessions/session-entry-provenance.ts"
need_file "$OPENCLAW/src/plugins/hook-types.ts"
need_text "$OPENCLAW/src/gateway/server-methods/client-types.ts" 'authenticatedUserProfile?'
need_text "$OPENCLAW/src/gateway/server-methods/client-types.ts" 'profileId: string'
need_text "$OPENCLAW/src/plugins/plugin-api.types.ts" 'profileAccess?: "independent" | "required"'
need_text "$OPENCLAW/src/plugins/runtime/types-core.ts" 'getSessionEntry:'
need_text "$OPENCLAW/src/config/sessions/session-entry-provenance.ts" 'source: "profile" | "channel" | "unknown"'
need_text "$OPENCLAW/src/plugins/hook-types.ts" '"before_prompt_build"'
need_text "$OPENCLAW/src/plugins/hook-types.ts" '"agent_end"'
echo "PASS OpenClaw durable-profile/session/hook contracts"

need_file "$WEKNORA/internal/middleware/auth.go"
need_file "$WEKNORA/internal/types/tenant_api_key.go"
need_text "$WEKNORA/internal/middleware/auth.go" 'X-External-User-ID'
need_text "$WEKNORA/internal/middleware/auth.go" 'RequireDirectHeader'
need_text "$WEKNORA/internal/types/tenant_api_key.go" 'KnowledgeBaseIDs'
need_text "$WEKNORA/internal/types/tenant_api_key.go" 'Capabilities'
echo "PASS WeKnora external-principal/API-key-scope contracts"

need_file "$OPENVIKING/openviking/server/auth/plugins/trusted.py"
need_file "$OPENVIKING/openviking/server/routers/skills.py"
need_file "$OPENVIKING/openviking/server/routers/sessions.py"
need_file "$OPENVIKING/openviking/server/routers/search.py"
need_text "$OPENVIKING/openviking/server/auth/plugins/trusted.py" 'X-OpenViking-Account'
need_text "$OPENVIKING/openviking/server/auth/plugins/trusted.py" 'X-OpenViking-User'
need_text "$OPENVIKING/openviking/server/routers/skills.py" '@router.post("/find")'
need_text "$OPENVIKING/openviking/server/routers/sessions.py" 'prefix="/api/v1/sessions"'
echo "PASS OpenViking trusted-principal/memory/skill contracts"
echo "PASS v0.6 upstream compatibility gate"
