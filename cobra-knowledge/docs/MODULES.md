# v0.4 模块与代码职责

## `internal/model`

核心数据契约：

- `GraphSnapshot / Entity / Relation / EvidenceRef`
- `Ontology / OntologyClass / DataProperty / ObjectRelation / Constraint / ReviewItem`
- `Assertion / RetrievalPlan / ContextPack`
- `GraphView / GraphViewNode / GraphViewEdge / GraphViewMeta`
- v0.4 新增 `OntologyManifest / OntologyVersionMeta / OntologyBinding / OntologyResolution / RegistryAuditEvent`

`GraphView` 仍是 WeKnora Overlay 唯一依赖的显示契约。Registry 的存储结构不会泄漏到前端。

## `internal/ontology`

- `discovery.go`：实体图模式统计与候选本体发现；
- `validator.go`：确定性本体和实体图校验；
- `compiler.go`：Ontology -> WeKnora ExtractConfig；
- `release.go`：候选本体审核完成后生成新的 approved 快照，不原地修改候选版本；
- `registry.go`：v0.4 正式 Registry。版本不可变、发布、active 指针、KB binding、rollback、checksum、audit。

Registry 对外只暴露逻辑 ID 和版本，不暴露 ontology 文件路径。默认 `FSRegistry` 可以后续替换 PostgreSQL Registry。

## `internal/graphview`

- `neo4j_http.go`：只读 WeKnora GraphRAG Neo4j 实体图；
- `ontology.go`：Ontology -> GraphView；
- `source.go`：`RegistryOntologySource`，通过 Registry 解析当前 KB 的 active/pinned published ontology。

v0.3 的 `FileOntologyBindings` 已退出运行链路。

## `internal/httpapi`

两类接口共享进程但权限边界分离：

### WeKnora-facing Graph API

```text
GET /api/v1/knowledge-bases/{kb_id}/graph?view=entity|ontology
```

沿用 WeKnora RBAC，不改变 v0.3 前端调用契约。

### Ontology Registry Governance API

```text
GET/POST/PUT /api/v1/registry/...
```

使用独立 `X-Cobra-Admin-Token`。普通 WeKnora Token 不能发布或回滚本体。

## `internal/mcp`

Agent 生产入口仍保持少量稳定工具：

- `context.retrieve`
- `context.get_evidence`

v0.4 支持通过 `COBRA_ONTOLOGY_REGISTRY_ROOT + COBRA_ONTOLOGY_KB_ID` 加载 Registry 当前正式本体。`COBRA_ONTOLOGY_FILE` 仅作为本地开发兼容方式。

## `internal/retrieval`

- `catalog.go`：Ontology -> Semantic Catalog；
- `planner.go`：最小充分 Retrieval Plan；
- `arbiter.go`：时效、来源、范围、版本和冲突裁决。

## `internal/graph`

- `resolver.go`：保守式实体归一；
- `assertion.go`：事实主张和 Evidence 构建。

## `integrations/weknora`

v0.4 **没有新增 WeKnora Overlay 改动**。仍使用 v0.3 的：

- `cobra-knowledge.ts`
- `GraphExplorer.vue`
- `0001-add-ontology-graph-switcher.patch`
- `apply-overlay.sh`

因此 WeKnora 上游冲突面没有随 v0.4 扩大。
