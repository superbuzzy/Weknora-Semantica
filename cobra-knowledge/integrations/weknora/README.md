# WeKnora graph-view overlay

This integration keeps the WeKnora upstream checkout clean. It adds one reusable graph viewer to the existing knowledge-graph settings area and exposes two modes on the same canvas:

- **实体图**: reads the current WeKnora GraphRAG graph directly from its Neo4j database through CobraKnowledge Graph API.
- **本体图**: resolves the active/pinned published ontology from CobraKnowledge Registry and projects classes, data properties, class hierarchy and object relations into the same graph-view contract.

## Apply without modifying upstream

```bash
./cobra-knowledge/integrations/weknora/apply-overlay.sh \
  ./upstream/weknora \
  ./build/weknora-v0.4
```

Run or build WeKnora from `build/weknora-v0.4`. The `upstream/weknora` checkout remains untouched, so a future upstream update is simply:

```bash
cd upstream/weknora
git pull
cd ../..
./cobra-knowledge/integrations/weknora/apply-overlay.sh upstream/weknora build/weknora-v0.4
```

If the small `GraphSettings.vue` patch no longer applies, the script fails instead of silently producing a broken UI. The two added files are copied as overlay files and do not conflict with upstream unless WeKnora later creates files with the same names.

## Reverse proxy

Prefer same-origin deployment. Example Nginx location:

```nginx
location /cobra-knowledge/ {
    proxy_pass http://cobra-graph-api:8090/;
    proxy_set_header Authorization $http_authorization;
    proxy_set_header X-Tenant-ID $http_x_tenant_id;
    proxy_set_header Accept-Language $http_accept_language;
}
```

The frontend defaults to `/cobra-knowledge`. To use a different base URL, set `VITE_COBRA_KNOWLEDGE_API_BASE` when building the derived WeKnora frontend.

## Authorization

In production, run Graph API with `COBRA_GRAPH_AUTH_MODE=weknora` and configure `COBRA_WEKNORA_BASE_URL`. CobraKnowledge delegates access checks to WeKnora's existing `GET /api/v1/knowledge-bases/:id` endpoint using the caller's Authorization and tenant headers. This avoids duplicating WeKnora RBAC.


## v0.4 compatibility note

v0.4 does not add any new WeKnora patch or overlay file. The exact v0.3 UI integration is reused. Ontology file-binding resolution was replaced entirely inside CobraKnowledge with Ontology Registry, so the WeKnora upstream conflict surface remains unchanged.
