# CobraKnowledge v0.3

CobraKnowledge is a Go-native enterprise Agent retrieval and context-governance core. WeKnora is a replaceable knowledge source, Semantica is a design reference only, and the CobraKnowledge runtime imports neither upstream project.

## Core architecture

```text
                           Agent Runtime
                                │
                              Skill
                                │
                         Context MCP / API
                                │
                    Retrieval Control Plane
                 Resolver -> Planner -> Retrievers
                           -> Arbiter -> Assembler
                                │
                            Context Pack

Knowledge assets:
  Wiki Graph        Entity Graph        Ontology Graph        Business Data
  WeKnora           WeKnora Neo4j       CobraKnowledge        MCP / API
```

The three graphs are knowledge assets. The core capability is the retrieval strategy: what to search, where to search, which evidence is trustworthy, how conflicts and freshness are handled, and when retrieval should stop.

## v0.3: ontology graph visible in WeKnora

v0.3 adds a stable graph-visualization boundary without merging the WeKnora and CobraKnowledge storage models.

- `cobra-graph-api` exposes one graph contract with `view=entity|ontology`.
- **Entity graph** is read-only from WeKnora's existing Neo4j GraphRAG storage.
- **Ontology graph** is projected from the ontology bound to the current WeKnora knowledge base.
- The WeKnora overlay adds one `GraphExplorer` in the existing graph settings area with **实体图 / 本体图** switch buttons.
- Wiki graph remains on the official WeKnora Wiki Browser.
- Graph API authorization is delegated back to WeKnora RBAC.
- WeKnora upstream remains clean: the integration is applied to a derived build tree.

See `docs/GRAPH_VISUALIZATION.md` for deployment and UI details.

## Existing v0.2 core

v0.3 keeps the v0.2 retrieval and ontology core:

- native bottom-up ontology discovery from entity graph patterns;
- stable machine IDs independent of Chinese naming;
- conservative entity resolution;
- Assertion/Evidence fact layer;
- deterministic ontology and graph validation;
- semantic catalog + retrieval planner;
- freshness/source/time-aware Knowledge Arbiter;
- concurrent Context Service + Context Pack;
- WeKnora RAG/Chunk adapters;
- MCP stdio server and domain Skills.

## Quick start

```bash
make test
make vet
make build
```

Bootstrap a candidate ontology from a WeKnora `GraphData` export:

```bash
bin/cobra-knowledge bootstrap \
  -in examples/weknora-graph.json \
  -domain distribution_network \
  -out out/bootstrap
```

## Graph API

Create a KB-to-ontology binding file from `configs/ontology-bindings.example.json`, then:

```bash
export COBRA_NEO4J_URL=http://neo4j:7474
export COBRA_NEO4J_USER=neo4j
export COBRA_NEO4J_PASSWORD='***'
export COBRA_ONTOLOGY_BINDINGS=/app/configs/ontology-bindings.json
export COBRA_WEKNORA_BASE_URL=http://weknora:8080
export COBRA_GRAPH_AUTH_MODE=weknora

bin/cobra-graph-api -listen :8090
```

Endpoints:

```text
GET /healthz
GET /api/v1/knowledge-bases/{kb_id}/graph?view=entity&limit=160
GET /api/v1/knowledge-bases/{kb_id}/graph?view=ontology
```

## WeKnora UI overlay

```bash
./integrations/weknora/apply-overlay.sh \
  ../upstream/weknora \
  ../build/weknora-v0.3
```

Build/run WeKnora from the derived directory. The upstream checkout stays untouched and can continue to `git pull` normally.

Prefer a same-origin reverse proxy:

```nginx
location /cobra-knowledge/ {
    proxy_pass http://cobra-graph-api:8090/;
    proxy_set_header Authorization $http_authorization;
    proxy_set_header X-Tenant-ID $http_x_tenant_id;
    proxy_set_header Accept-Language $http_accept_language;
}
```

## MCP runtime

The Agent-facing path is unchanged:

```bash
export COBRA_ONTOLOGY_FILE=/path/to/ontology.json
export COBRA_ENTITY_GRAPH_FILE=/path/to/normalized-graph.json
export WEKNORA_BASE_URL=http://localhost:8080
export WEKNORA_API_KEY=sk-xxxxx
export WEKNORA_KB_IDS=kb-1,kb-2

go run ./cmd/context-mcp
```

Production-facing tools remain intentionally small:

- `context.retrieve`
- `context.get_evidence`

Fine-grained ontology/retrieval/arbitration tools remain for governance and audit.

## Upstream isolation

CobraKnowledge never imports packages from WeKnora or Semantica. WeKnora integration is through HTTP/Neo4j read adapters and a derived frontend overlay. Semantica remains research input only. Upstream updates are absorbed at adapter/overlay boundaries instead of long-lived forks.
