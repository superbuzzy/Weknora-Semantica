# CobraKnowledge v0.2

CobraKnowledge v0.2 is a Go-native enterprise knowledge retrieval and ontology control plane. It treats WeKnora as a replaceable knowledge source and Semantica as a design reference only. Neither upstream project is modified or required at runtime.

## Core idea

```text
Wiki Graph        Entity Graph        Ontology Graph        Business Data
    \                 |                  /                     /
     \                |                 /                     /
              Retrieval Control Plane
       Semantic Resolver -> Planner -> Retrievers
                         -> Arbiter -> Context Assembler
                                  |
                              Context Pack
                                  |
                          Agent Runtime / Skill
```

The three graphs are knowledge assets. The core product is the retrieval strategy: what to search, where to search, how much evidence is enough, which source wins conflicts, and when retrieval should stop.

## v0.2 changes

- Rewritten from Python to Go 1.23, standard library only.
- Removed Semantica runtime/import dependency.
- Added native bottom-up ontology discovery from entity graph patterns.
- Added stable machine IDs independent of Chinese-to-English naming.
- Added conservative entity resolution and property-conflict preservation.
- Added Assertion/Evidence fact layer.
- Added deterministic ontology/graph validation.
- Added fast retrieval planner and semantic catalog compiler.
- Added freshness/source/time conflict arbiter.
- Added concurrent Context Service and Context Pack assembler.
- Added WeKnora RAG HTTP adapter and Chunk evidence resolver.
- Added MCP stdio server with `context.retrieve` as the main Agent-facing entry.
- Rebuilt ontology/retrieval Skills and prompts around the v0.2 contracts.

## Quick start

```bash
go test ./...
go build ./cmd/cobra
go build ./cmd/context-mcp

# Full bootstrap from a WeKnora GraphData export
go run ./cmd/cobra bootstrap \
  -in examples/weknora-graph.json \
  -domain distribution_network \
  -out out/bootstrap

# Plan a query from the generated ontology
go run ./cmd/cobra plan \
  -ontology out/bootstrap/candidate-ontology.json \
  -query '金牛线为什么转供能力不足？' \
  -out out/retrieval-plan.json
```

## MCP runtime

Configure available sources with environment variables:

```bash
export COBRA_ONTOLOGY_FILE=/path/to/ontology.json
export COBRA_ENTITY_GRAPH_FILE=/path/to/normalized-graph.json
export WEKNORA_BASE_URL=http://localhost:8080
export WEKNORA_API_KEY=sk-xxxxx
export WEKNORA_KB_IDS=kb-1,kb-2

go run ./cmd/context-mcp
```

Main tools:

- `context.retrieve`
- `context.get_evidence`
- `retrieval.plan` (development/audit)
- `ontology.discover` (development/governance)
- `ontology.compile_weknora` (development/governance)
- `knowledge.arbitrate` (development/audit)

## Upstream isolation

The workspace keeps upstream code under `../upstream/` for research only. CobraKnowledge does not import packages from either upstream repository. Upstream changes are absorbed through Adapter contracts.

See `ARCHITECTURE.md` and `docs/MODULES.md` for implementation details.

### Domain semantic/policy overlay

Ontology discovery intentionally does not invent domain synonyms. Put expert/Skill-maintained aliases and preferred sources in a separate overlay:

```bash
go run ./cmd/cobra plan \
  -ontology out/bootstrap/candidate-ontology.json \
  -overlay configs/semantic-overlay.json \
  -query '金牛线为什么转供能力不足？'
```

For MCP runtime, set `COBRA_CATALOG_OVERLAY_FILE`. This keeps business policy separate from core code and from auto-discovered ontology candidates.

### Upstream update discipline

Run `scripts/check-upstream-clean.sh` before integrating a new upstream snapshot. If the upstream directories are real git clones, `scripts/sync-upstreams.sh` performs a fast-forward-only update. CobraKnowledge never writes into either upstream tree.
