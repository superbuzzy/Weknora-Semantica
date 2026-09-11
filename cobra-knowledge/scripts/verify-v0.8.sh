#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "[1/10] version"
[ "$(cat VERSION)" = "0.8.0" ]

echo "[2/10] Go test/vet/build"
go test ./... -count=1 -timeout=45s
go vet ./...
mkdir -p bin
go build -o bin/cobra-knowledge ./cmd/cobra
go build -o bin/cobra-context-mcp ./cmd/context-mcp
go build -o bin/cobra-graph-api ./cmd/graph-api

echo "[3/10] JavaScript syntax"
for f in \
  integrations/openclaw/workspace-core/index.js \
  integrations/openclaw/knowledge-plugin/index.js \
  integrations/openclaw/knowledge-plugin/lib/*.js \
  integrations/openclaw/knowledge-plugin/browser/index.js \
  integrations/openclaw/openviking-plugin/index.js \
  integrations/openclaw/openviking-plugin/lib/*.js \
  integrations/openclaw/openviking-plugin/browser/index.js; do
  node --check "$f"
done

echo "[4/10] Workspace / Knowledge / OpenViking contracts"
node --test integrations/openclaw/workspace-core/test/*.test.js
node --test integrations/openclaw/knowledge-plugin/test/*.test.js
node --test integrations/openclaw/openviking-plugin/test/*.test.js

echo "[5/10] one canonical Workspace resolver"
if find integrations/openclaw/knowledge-plugin integrations/openclaw/openviking-plugin -path "*/workspace-registry.js" -print -quit | grep -q .; then
  echo "FAIL workspace registry must have one canonical implementation" >&2
  exit 1
fi

echo "[6/10] browser cannot address downstream credentials/trusted headers"
if grep -R -E 'X-API-Key|X-Tenant-ID|X-External-User-ID|X-OpenViking-(Account|User)|LEECLAW_RUNTIME_TOKEN|WEKNORA_API_KEY|OPENVIKING_API_KEY' integrations/openclaw/*-plugin/browser; then
  echo "FAIL browser knows downstream credentials or trusted headers" >&2
  exit 1
fi

echo "[7/10] obsolete static identity / Graph API abstractions are absent"
if grep -R -E 'weknoraBearerToken|forwardAuthenticatedUser|params\.tenantId|params\.accountId|params\.userId|params\.externalUserId|graphApiBaseUrl' \
  integrations/openclaw/knowledge-plugin/index.js \
  integrations/openclaw/knowledge-plugin/lib \
  integrations/openclaw/knowledge-plugin/openclaw.plugin.json \
  integrations/openclaw/openviking-plugin/index.js \
  integrations/openclaw/openviking-plugin/lib \
  integrations/openclaw/openviking-plugin/openclaw.plugin.json \
  configs/openclaw-v0.8.example.json; then
  echo "FAIL legacy identity/workspace/Graph API path remains" >&2
  exit 1
fi

echo "[8/10] Agent Knowledge tools are host-declared and Skill Runtime uses host authority"
grep -Fq '"leeclaw_context_retrieve"' integrations/openclaw/knowledge-plugin/openclaw.plugin.json
grep -Fq '"leeclaw_context_get_evidence"' integrations/openclaw/knowledge-plugin/openclaw.plugin.json
grep -Fq 'api.registerTool' integrations/openclaw/knowledge-plugin/index.js
grep -Fq 'requiresToolAuthority: true' integrations/openclaw/openviking-plugin/lib/agent-runtime.js
grep -Fq 'toolsAllow' integrations/openclaw/openviking-plugin/lib/agent-runtime.js
grep -Fq 'toolAuthority?.allows' integrations/openclaw/openviking-plugin/lib/skill-runtime.js
grep -Fq 'sessionId' integrations/openclaw/workspace-core/index.js
grep -Fq 'knowledge_base_ids' integrations/openclaw/knowledge-plugin/lib/context-runtime-client.js
grep -Fq 'ErrEvidenceOutsideScope' internal/runtimecontext/service.go
grep -Fq 'candidateSourceRank' integrations/openclaw/openviking-plugin/lib/skill-runtime.js

echo "[9/10] v0.7 active config/gate files were replaced"
for obsolete in \
  configs/openclaw-v0.7.example.json \
  configs/workspaces-v0.7.example.json \
  compatibility/upstreams-v0.7.json \
  scripts/verify-v0.7.sh \
  scripts/check-v0.7-upstreams.sh; do
  [ ! -e "$obsolete" ] || { echo "FAIL obsolete active file remains: $obsolete" >&2; exit 1; }
done

echo "[10/10] runtime server does not expose trusted service headers to browser CORS"
grep -Fq 'Access-Control-Allow-Headers", "Content-Type"' internal/httpapi/server.go
if grep -F 'Access-Control-Allow-Headers' internal/httpapi/server.go | grep -Eq 'LeeClaw|Tenant|API-Key|OpenViking'; then
  echo "FAIL trusted runtime headers exposed through CORS" >&2
  exit 1
fi

echo "PASS v0.8 local verification"
