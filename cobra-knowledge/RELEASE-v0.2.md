# CobraKnowledge v0.2 Release

## Release position

v0.2 is the first self-owned CobraKnowledge core. It replaces the v0.1 Python/Semantica bridge with a Go-native implementation and makes both WeKnora and Semantica replaceable upstream references.

## Verified behaviors

- WeKnora GraphData -> normalized graph -> entity resolution -> candidate ontology -> WeKnora ExtractConfig.
- Chinese business labels remain first-class; machine IDs are deterministic hashes.
- Planner can resolve expert-maintained aliases through Semantic Overlay.
- Entity Retriever consumes planner target IDs and returns only relevant facts.
- Relation provenance from WeKnora remains marked `derived` because upstream relationships do not persist relation-level chunk IDs.
- Arbiter supports effective-time selection, freshness expiry, source priority and unresolved conflict.
- MCP initialize/tools-list/context.retrieve smoke test passes.
- No Go runtime file imports or references Semantica.
- Upstream source snapshots are untouched by CobraKnowledge build/test/demo.

## Recommended next release

v0.3 should focus on a governed persistent graph store, incremental synchronization, ontology review API, source-binding execution and evaluation datasets. Do not expand into a large ontology UI before the governed graph lifecycle is proven on real domain documents.
