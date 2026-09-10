# Changelog

## v0.3.0

### Graph visualization
- Added a stable `GraphView` contract shared by entity and ontology visualization.
- Added ontology graph projection for classes, data properties, hierarchy and object relations.
- Added read-only WeKnora Neo4j entity graph source through the Neo4j transactional HTTP endpoint.
- Added KB-scoped graph HTTP API with `view=entity|ontology`.
- Added knowledge-base-to-ontology external binding configuration.

### WeKnora UI integration
- Added a derived WeKnora overlay with a reusable `GraphExplorer.vue`.
- Added an `实体图 / 本体图` switch in the existing graph settings area.
- Kept Wiki graph on the existing WeKnora Wiki Browser.
- Added overlay application script that never modifies the upstream checkout.

### Security and compatibility
- Delegated graph-view authorization to WeKnora RBAC using the caller's Authorization and tenant headers.
- Kept WeKnora Neo4j schema read-only and unchanged.
- Kept Semantica as a research reference only; no runtime dependency.


## v0.2.0

### Architecture
- Replaced the v0.1 Python extension with a Go-native CobraKnowledge Core.
- Removed all Semantica runtime/import dependencies; Semantica remains a research reference only.
- Kept WeKnora behind Adapter contracts; no upstream source modifications.

### Ontology
- Native bottom-up graph pattern analysis and candidate ontology generation.
- Stable machine IDs independent of Chinese naming conventions.
- Domain/Range confidence, review queue, deterministic ontology/graph validation.
- Immutable approved-version registry and WeKnora ExtractConfig compiler.
- Optional semantic induction interface for LLM-assisted aliases/hierarchy/merge suggestions.

### Graph governance
- Conservative entity resolution with business-key priority.
- Property conflicts are preserved instead of overwritten.
- Assertion/Evidence model separates mutable facts from entity identity.

### Retrieval
- Semantic Catalog and domain overlay for aliases/source policies.
- Fast Planner with minimum-sufficient source selection and parallel groups.
- Deterministic Arbiter for valid time, freshness, source priority and conflicts.
- Concurrent Context Service and Context Pack assembly.

### Integration
- WeKnora knowledge-search and chunk evidence HTTP adapters.
- MCP stdio server with `context.retrieve` as the production-facing main entry.
- Rebuilt ontology and retrieval Skills and prompt contracts.
