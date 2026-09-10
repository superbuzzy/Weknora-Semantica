#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "[1/8] Go tests"; go test ./... -count=1 -timeout=30s
echo "[2/8] Go vet"; go vet ./...
echo "[3/8] Go build"; make build
echo "[4/8] JavaScript syntax"
while IFS= read -r file; do node --check "$file"; done < <(find integrations/openclaw -type f -name '*.js' | sort)
echo "[5/8] Adapter/principal/runtime contract tests"
node --test integrations/openclaw/knowledge-plugin/test/*.test.js
node --test integrations/openclaw/openviking-plugin/test/*.test.js
echo "[6/8] Browser cannot address upstream or identity fields"
if grep -R -n -E '/api/v1|WEKNORA_BASE_URL|OPENVIKING_BASE_URL|externalUserId|tenantId|accountId|userId' \
  integrations/openclaw/knowledge-plugin/browser integrations/openclaw/openviking-plugin/browser; then
  echo "FAIL browser bypasses Adapter/Principal boundary" >&2; exit 1
fi
echo "[7/8] Legacy identity configuration must be gone"
if grep -R -n -E 'weknoraBearerToken|forwardAuthenticatedUser|config\.externalUserId|config\.userId|config\.accountId|params\.externalUserId|params\.tenantId|params\.userId|params\.accountId' \
  integrations/openclaw/knowledge-plugin/index.js integrations/openclaw/knowledge-plugin/lib integrations/openclaw/knowledge-plugin/openclaw.plugin.json \
  integrations/openclaw/openviking-plugin/index.js integrations/openclaw/openviking-plugin/lib integrations/openclaw/openviking-plugin/openclaw.plugin.json \
  configs/openclaw-v0.6.example.json; then
  echo "FAIL legacy identity override remains" >&2; exit 1
fi
echo "[8/8] OpenClaw durable profile required"
grep -R -q 'authenticatedUserProfile.*profileId\|authenticatedUserProfile?.profileId' integrations/openclaw/knowledge-plugin integrations/openclaw/openviking-plugin

echo "PASS v0.6 local verification"
