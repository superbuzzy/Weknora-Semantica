#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "[1/6] Go tests"
go test ./... -count=1 -timeout=30s

echo "[2/6] Go vet"
go vet ./...

echo "[3/6] Go build"
make build

echo "[4/6] JavaScript syntax"
while IFS= read -r file; do
  node --check "$file"
done < <(find integrations/openclaw -type f -name '*.js' | sort)

echo "[5/6] Adapter contract tests"
node --test integrations/openclaw/knowledge-plugin/test/*.test.js
node --test integrations/openclaw/openviking-plugin/test/*.test.js

echo "[6/6] Browser isolation"
if grep -R -n -E '/api/v1|WEKNORA_BASE_URL|OPENVIKING_BASE_URL' \
  integrations/openclaw/knowledge-plugin/browser \
  integrations/openclaw/openviking-plugin/browser; then
  echo "FAIL browser code bypasses plugin Adapter/BFF" >&2
  exit 1
fi

echo "PASS v0.5 local verification"
