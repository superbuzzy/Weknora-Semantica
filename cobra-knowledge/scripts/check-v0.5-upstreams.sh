#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 3 ]; then
  echo "usage: $0 <openclaw-source> <weknora-source> <openviking-source>" >&2
  exit 2
fi

OPENCLAW="$1"
WEKNORA="$2"
OPENVIKING="$3"

need_file() {
  if [ ! -f "$1" ]; then
    echo "FAIL missing: $1" >&2
    exit 1
  fi
}

need_text() {
  local file="$1"
  local pattern="$2"
  if ! grep -Fq "$pattern" "$file"; then
    echo "FAIL contract '$pattern' missing in $file" >&2
    exit 1
  fi
}

need_file "$OPENCLAW/package.json"
need_file "$OPENCLAW/src/plugin-sdk/control-ui.ts"
need_file "$OPENCLAW/src/plugin-sdk/plugin-entry.ts"
need_file "$OPENCLAW/src/plugins/host-hooks.ts"
need_file "$OPENCLAW/src/plugins/plugin-api.types.ts"
need_file "$OPENCLAW/pnpm-workspace.yaml"
need_text "$OPENCLAW/package.json" '"version": "2026.9.3"'
need_text "$OPENCLAW/src/plugin-sdk/control-ui.ts" 'defineControlUiPlugin'
need_text "$OPENCLAW/src/plugins/plugin-api.types.ts" 'registerGatewayMethod'
need_text "$OPENCLAW/src/plugins/plugin-api.types.ts" 'registerControlUiDescriptor'
need_text "$OPENCLAW/pnpm-workspace.yaml" 'extensions/*'

echo "PASS OpenClaw 2026.9.3 plugin/control-ui contracts"

need_file "$WEKNORA/frontend/package.json"
need_file "$WEKNORA/frontend/src/api/knowledge-base/index.ts"
need_file "$WEKNORA/frontend/src/api/wiki/index.ts"
need_file "$WEKNORA/frontend/src/api/tenant/members.ts"
need_file "$WEKNORA/frontend/src/api/organization/index.ts"
need_file "$WEKNORA/internal/middleware/auth.go"
need_text "$WEKNORA/frontend/package.json" '"version": "0.8.0"'
need_text "$WEKNORA/frontend/src/api/knowledge-base/index.ts" '/api/v1/knowledge-bases'
need_text "$WEKNORA/frontend/src/api/knowledge-base/index.ts" '/activity'
need_text "$WEKNORA/frontend/src/api/wiki/index.ts" '/api/v1/knowledgebase/'
need_text "$WEKNORA/frontend/src/api/tenant/members.ts" '/members'
need_text "$WEKNORA/frontend/src/api/organization/index.ts" '/shares'
need_text "$WEKNORA/internal/middleware/auth.go" 'X-External-User-ID'
need_text "$WEKNORA/internal/middleware/auth.go" 'X-External-User-Token'

echo "PASS WeKnora 0.8.0 knowledge/RBAC/API-principal contracts"

need_file "$OPENVIKING/examples/openclaw-plugin/package.json"
need_file "$OPENVIKING/examples/openclaw-plugin/openclaw.plugin.json"
need_file "$OPENVIKING/openviking/server/routers/skills.py"
need_file "$OPENVIKING/openviking/server/routers/search.py"
need_file "$OPENVIKING/openviking/server/routers/sessions.py"
need_text "$OPENVIKING/examples/openclaw-plugin/package.json" '"version": "2026.6.18"'
need_text "$OPENVIKING/examples/openclaw-plugin/openclaw.plugin.json" '"kind": "context-engine"'
need_text "$OPENVIKING/openviking/server/routers/skills.py" '@router.post("/find")'
need_text "$OPENVIKING/openviking/server/routers/skills.py" '@router.get("/{skill_name}")'
need_text "$OPENVIKING/openviking/server/routers/search.py" 'class FindRequest'
need_text "$OPENVIKING/openviking/server/routers/sessions.py" 'prefix="/api/v1/sessions"'

echo "PASS OpenViking 2026.6.18 context/memory/skill contracts"
echo "PASS v0.5 upstream compatibility gate"
