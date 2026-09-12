#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"; cd "$ROOT"
echo "[1/9] version"; [ "$(cat VERSION)" = "0.10.0" ]
echo "[2/9] Go"; go test ./... -count=1 -timeout=45s; go vet ./...; mkdir -p bin; go build -o bin/cobra-knowledge ./cmd/cobra; go build -o bin/cobra-context-mcp ./cmd/context-mcp; go build -o bin/cobra-graph-api ./cmd/graph-api
echo "[3/9] JS"; for f in integrations/openclaw/workspace-core/index.js integrations/openclaw/knowledge-plugin/index.js integrations/openclaw/knowledge-plugin/lib/*.js integrations/openclaw/openviking-plugin/index.js integrations/openclaw/openviking-plugin/lib/*.js integrations/openclaw/promotion-plugin/index.js integrations/openclaw/promotion-plugin/lib/*.js; do node --check "$f"; done
echo "[4/9] existing plugin tests"; node --test integrations/openclaw/workspace-core/test/*.test.js; node --test integrations/openclaw/knowledge-plugin/test/*.test.js; node --test integrations/openclaw/openviking-plugin/test/*.test.js
echo "[5/9] Promotion contracts"; grep -Fq 'type Candidate struct' internal/promotion/types.go; grep -Fq 'func promotionGate' internal/promotion/service.go; grep -Fq 'verifyTraceability' internal/promotion/openviking_backend.go; grep -Fq 'delete_created_knowledge' internal/promotion/weknora_backend.go; grep -Fq 'restore_previous_skill' internal/promotion/openviking_backend.go
echo "[6/9] trust boundary"; grep -Fq 'LEECLAW_PROMOTION_TOKEN' cmd/graph-api/main.go; grep -Fq 'promotion authentication required' internal/promotionapi/handler.go; grep -Fq 'operator.admin' integrations/openclaw/promotion-plugin/index.js
echo "[7/9] Agent cannot publish"; grep -Fq 'leeclaw_context_retrieve' integrations/openclaw/knowledge-plugin/openclaw.plugin.json; ! grep -Eq 'promotion.*(publish|rollback)' integrations/openclaw/knowledge-plugin/openclaw.plugin.json
echo "[8/9] active files"; grep -Fq 'verify-v0.10.sh' Makefile; for f in configs/openclaw-v0.9.example.json configs/workspaces-v0.9.example.json compatibility/upstreams-v0.9.json scripts/verify-v0.9.sh scripts/check-v0.9-upstreams.sh; do [ ! -e "$f" ] || { echo "FAIL obsolete active file: $f"; exit 1; }; done
echo "[9/9] upstream clean"; ./scripts/check-upstream-clean.sh
echo "PASS v0.10 local verification"
