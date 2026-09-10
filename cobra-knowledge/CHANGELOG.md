# Changelog

## v0.4.0

- 新增正式 `ontology.Registry` 抽象与默认 `FSRegistry`。
- 本体版本注册后不可变，并使用 SHA-256 校验内容完整性。
- 发布状态与本体 payload 分离，新增 `active_version` 指针。
- KB Binding 支持 `active` 与 `pinned`。
- 支持历史 published 版本重新激活，实现指针级回滚。
- 新增 Registry append-only 审计事件。
- 新增 `approve-ontology`、`registry-register`、`registry-publish`、`registry-activate`、`registry-bind`、`registry-resolve` CLI。
- Graph API 的 ontology 视图由文件路径绑定切换到 Registry 解析；原 WeKnora 图谱 GET 契约不变。
- 新增 Registry REST 治理 API，并与 WeKnora 用户身份分离为独立管理员 Token。
- Context MCP 支持从 Registry 获取正式本体。
- v0.4 不新增任何 WeKnora Overlay 或后端修改。

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
