#!/usr/bin/env bash
set -euo pipefail
[ "$#" -eq 3 ] || { echo "usage: $0 <openclaw> <weknora> <openviking>" >&2; exit 2; }
OC="$1"; WK="$2"; OV="$3"
need(){ grep -R -Fq "$2" "$1" || { echo "missing upstream contract: $2 in $1" >&2; exit 1; }; }
need "$OC/src/plugins" 'profileAccess?: "independent" | "required"'
need "$OC/src/gateway/server-methods" '"users.list"'
need "$OC/src" 'authenticatedUserProfile?.profileId'
need "$OC/src" 'createdActor'
need "$OC/src/plugins" 'before_prompt_build'
need "$OC/src/plugin-sdk" 'defineControlUiPlugin'
need "$WK/internal/router" '/knowledge-bases/:id/knowledge'
need "$WK/internal/router" 'http.MethodPut, "/api/v1/knowledge-bases/:id"'
need "$WK/internal/router" 'http.MethodDelete, "/api/v1/knowledge-bases/:id"'
need "$WK/internal/router" '/knowledge-bases/:id/shares'
need "$WK/internal/router" '/knowledge-bases/:id/tags'
need "$WK/internal/router" '/knowledge-bases/:id/faq'
need "$WK/internal/middleware" 'X-External-User-ID'
need "$OV" '/api/v1/skills'
need "$OV" 'X-OpenViking-Account'
need "$OV" 'X-OpenViking-User'
need "$OV" '/api/v1/sessions'
echo "PASS v0.7 upstream compatibility gate"
