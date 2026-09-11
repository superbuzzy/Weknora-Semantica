#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "[1/12] version"
[ "$(cat VERSION)" = "0.9.0" ]

echo "[2/12] Go test/vet/build"
go test ./... -count=1 -timeout=45s
go vet ./...
mkdir -p bin
go build -o bin/cobra-knowledge ./cmd/cobra
go build -o bin/cobra-context-mcp ./cmd/context-mcp
go build -o bin/cobra-graph-api ./cmd/graph-api

echo "[3/12] JavaScript syntax"
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

echo "[4/12] Workspace / Knowledge / OpenViking contracts"
node --test integrations/openclaw/workspace-core/test/*.test.js
node --test integrations/openclaw/knowledge-plugin/test/*.test.js
node --test integrations/openclaw/openviking-plugin/test/*.test.js

echo "[5/12] retrieval governance contracts"
grep -Fq 'RequiredSources' internal/model/retrieval.go
grep -Fq 'BlockingGaps' internal/model/retrieval.go
grep -Fq 'SourceStatus' internal/model/retrieval.go
grep -Fq 'FederateCatalogs' internal/retrieval/catalog.go
grep -Fq 'BusinessQuerySource' internal/context/business_retriever.go
grep -Fq 'EntityQuerySource' internal/context/runtime_retrievers.go
grep -Fq 'EntityQuery(' internal/graphview/entity_query.go

echo "[6/12] required-source completeness gate"
grep -Fq 'result.Required && !result.Satisfied' internal/context/assembler.go
grep -Fq 'unresolved fact conflict' internal/context/assembler.go
grep -Fq 'plan.BlockingIssues' internal/context/assembler.go

echo "[7/12] Evidence v2"
grep -Fq 'SourceType' internal/model/graph.go
grep -Fq 'EffectiveAt' internal/model/graph.go
grep -Fq 'Provenance' internal/model/graph.go

echo "[8/12] one canonical Workspace resolver"
if find integrations/openclaw/knowledge-plugin integrations/openclaw/openviking-plugin -path "*/workspace-registry.js" -print -quit | grep -q .; then
  echo "FAIL workspace registry must have one canonical implementation" >&2
  exit 1
fi

echo "[9/12] browser cannot address trusted downstream identity"
if grep -R -E 'X-API-Key|X-Tenant-ID|X-External-User-ID|X-OpenViking-(Account|User)|LEECLAW_RUNTIME_TOKEN|WEKNORA_API_KEY|OPENVIKING_API_KEY' integrations/openclaw/*-plugin/browser; then
  echo "FAIL browser knows downstream credentials or trusted headers" >&2
  exit 1
fi

echo "[10/12] Agent Knowledge and Skill authority"
grep -Fq '"leeclaw_context_retrieve"' integrations/openclaw/knowledge-plugin/openclaw.plugin.json
grep -Fq '"leeclaw_context_get_evidence"' integrations/openclaw/knowledge-plugin/openclaw.plugin.json
grep -Fq 'requiresToolAuthority: true' integrations/openclaw/openviking-plugin/lib/agent-runtime.js
grep -Fq 'toolAuthority?.allows' integrations/openclaw/openviking-plugin/lib/skill-runtime.js
grep -Fq 'ErrEvidenceOutsideScope' internal/runtimecontext/service.go

echo "[11/12] obsolete active v0.8 config/gates were replaced"
for obsolete in \
  configs/openclaw-v0.8.example.json \
  configs/workspaces-v0.8.example.json \
  compatibility/upstreams-v0.8.json \
  scripts/verify-v0.8.sh \
  scripts/check-v0.8-upstreams.sh; do
  [ ! -e "$obsolete" ] || { echo "FAIL obsolete active file remains: $obsolete" >&2; exit 1; }
done

echo "[12/12] upstream source trees remain clean"
./scripts/check-upstream-clean.sh

echo "PASS v0.9 local verification"
