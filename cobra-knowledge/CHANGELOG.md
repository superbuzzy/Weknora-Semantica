# Changelog

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
