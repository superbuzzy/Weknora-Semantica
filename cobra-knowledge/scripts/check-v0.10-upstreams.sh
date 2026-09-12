#!/usr/bin/env bash
set -euo pipefail
[ "$#" -eq 3 ] || { echo "usage: $0 <openclaw-source> <weknora-source> <openviking-source>" >&2; exit 2; }
OPENCLAW="$1"; WEKNORA="$2"; OPENVIKING="$3"
need(){ grep -R -Fq "$2" "$1" || { echo "FAIL contract '$2' missing in $1" >&2; exit 1; }; }
need "$OPENCLAW/src" "authenticatedUserProfile"
need "$OPENCLAW/src" "registerGatewayMethod"
need "$OPENCLAW/src" "registerTool"
need "$OPENCLAW/src" "toolsAllow"
need "$WEKNORA/internal" "X-External-User-ID"
need "$WEKNORA/internal" "knowledge-search"
need "$WEKNORA/internal" "chunks/by-id"
need "$WEKNORA/internal" "knowledge/manual"
need "$OPENVIKING/openviking/server" "X-OpenViking-Account"
need "$OPENVIKING/openviking/server" "X-OpenViking-User"
need "$OPENVIKING/openviking/server/routers/skills.py" '@router.post("/validate")'
need "$OPENVIKING/openviking/server/routers/skills.py" '@router.put("/{skill_name}")'
need "$OPENVIKING/openviking/server/routers/skills.py" '@router.delete("/{skill_name}")'
need "$OPENVIKING/openviking/server/routers/resources.py" '@router.post("/skills")'
need "$OPENVIKING/openviking/server/routers/agent_evolution.py" 'experience_uri'
echo "PASS v0.10 upstream compatibility gate"
