# v0.3 模块与代码职责

## `internal/model`

核心数据契约：

- `GraphSnapshot / Entity / Relation / EvidenceRef`
- `Ontology / OntologyClass / DataProperty / ObjectRelation / Constraint / ReviewItem`
- `Assertion / RetrievalPlan / ContextPack`
- `GraphView / GraphViewNode / GraphViewEdge / GraphViewMeta`

`GraphView` 是 v0.3 新增的显示层稳定契约，UI 不直接消费 Neo4j 或 Ontology 存储结构。

## `internal/adapters/weknora`

### `graph.go`

兼容 WeKnora `GraphData -> GraphNode/GraphRelation` JSON，供离线本体发现和事实构建使用。

### `search.go`

调用 WeKnora 公共检索/Chunk API，为 Context Service 提供 RAG 与证据回溯。

## `internal/ontology`

- `discovery.go`：图模式统计和候选本体生成；
- `validator.go`：确定性本体/实体图校验；
- `compiler.go`：Ontology -> WeKnora ExtractConfig；
- `registry.go`：候选/正式不可变版本；
- v0.3 的可视化投影放在 `internal/graphview`，避免 Registry 与 UI 耦合。

## `internal/graph`

- `resolver.go`：保守实体归一；
- `assertion.go`：实体属性和关系 -> Assertion/Evidence。

## `internal/retrieval`

- `catalog.go`：Ontology -> Semantic Catalog；
- `planner.go`：最小充分检索规划；
- `arbiter.go`：时效、来源、冲突裁决。

## `internal/context`

并发执行 Retriever，统一收集结果、裁决并装配 Context Pack。

## `internal/mcp`

Agent Runtime 的 stdio MCP 入口。生产主入口继续是 `context.retrieve` / `context.get_evidence`。

## `internal/graphview`（v0.3）

### `ontology.go`

将正式或候选 Ontology 投影成 `GraphView`：Class/Property 节点、继承边、属性边、对象关系边。

### `neo4j_http.go`

通过 Neo4j transactional HTTP API 只读 WeKnora GraphRAG 图。复用 WeKnora 当前 `ENTITY<kb_id>` Label 规则，但不 import WeKnora 代码、不修改 Neo4j Schema。

### `source.go`

定义 `EntitySource` / `OntologySource` 接口，并通过外部 binding 文件将 WeKnora KB 映射到 Ontology 版本。

## `internal/access`（v0.3）

`WeKnoraAccessChecker` 将 Graph API 的当前用户身份头转交 WeKnora 标准 KB 读取 API，以 WeKnora 原 RBAC 为唯一权限判定依据。

## `internal/httpapi`（v0.3）

统一图谱 REST API：

```text
GET /api/v1/knowledge-bases/{kb_id}/graph?view=entity|ontology
```

负责视图路由、limit、超时、授权和 JSON 输出，不承载图谱业务模型。

## `cmd/graph-api`（v0.3）

独立 Graph API 进程。配置 Neo4j、Ontology binding、WeKnora RBAC 委托与监听地址。

## `integrations/weknora`（v0.3）

### `overlay/`

两个新增前端文件：

- `frontend/src/api/cobra-knowledge.ts`
- `frontend/src/views/knowledge/settings/GraphExplorer.vue`

### `patches/`

仅修改 WeKnora `GraphSettings.vue` 的小补丁：插入 GraphExplorer 并 import 组件。

### `apply-overlay.sh`

从干净上游复制出派生构建树，再应用 overlay 和 patch；不修改 upstream checkout。
