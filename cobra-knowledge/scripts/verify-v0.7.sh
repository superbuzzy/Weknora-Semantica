#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
[ "$(cat VERSION)" = "0.7.0" ]
go test ./... -count=1 -timeout=30s
go vet ./...
mkdir -p bin
go build -o bin/cobra-knowledge ./cmd/cobra
go build -o bin/cobra-context-mcp ./cmd/context-mcp
go build -o bin/cobra-graph-api ./cmd/graph-api
for f in integrations/openclaw/workspace-core/index.js integrations/openclaw/knowledge-plugin/index.js integrations/openclaw/knowledge-plugin/lib/*.js integrations/openclaw/knowledge-plugin/browser/index.js integrations/openclaw/openviking-plugin/index.js integrations/openclaw/openviking-plugin/lib/*.js integrations/openclaw/openviking-plugin/browser/index.js; do node --check "$f"; done
node --test integrations/openclaw/workspace-core/test/*.test.js
node --test integrations/openclaw/knowledge-plugin/test/*.test.js
node --test integrations/openclaw/openviking-plugin/test/*.test.js
if find integrations/openclaw/knowledge-plugin integrations/openclaw/openviking-plugin -path "*/workspace-registry.js" -print -quit | grep -q .; then
  echo "workspace registry must have one canonical implementation" >&2; exit 1
fi
if grep -R -E 'X-API-Key|X-Tenant-ID|X-OpenViking-(Account|User)|WEKNORA_API_KEY|OPENVIKING_API_KEY' integrations/openclaw/*-plugin/browser; then
  echo "browser must not know downstream credentials or trusted headers" >&2; exit 1
fi
if grep -R -E 'weknoraBearerToken|forwardAuthenticatedUser|params\.tenantId|params\.accountId|params\.userId|params\.externalUserId' integrations/openclaw/*-plugin --exclude-dir=test; then
  echo "legacy identity/workspace override detected" >&2; exit 1
fi
if grep -R -E 'weknoraApiKey|weknoraTenantId|(^|[^A-Za-z])workspaceId([^A-Za-z]|$)|(^|[^A-Za-z])accountId([^A-Za-z]|$)|(^|[^A-Za-z])userId([^A-Za-z]|$)' integrations/openclaw/*-plugin/openclaw.plugin.json; then
  echo "static downstream identity/workspace config must not return" >&2; exit 1
fi
echo "PASS v0.7 local verification"
