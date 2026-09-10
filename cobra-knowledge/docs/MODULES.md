# v0.2 模块与代码职责

## `internal/model`

唯一核心数据契约。上游 Adapter 必须转换成这些模型后才能进入业务层。

- `GraphSnapshot / Entity / Relation / EvidenceRef`
- `Ontology / OntologyClass / DataProperty / ObjectRelation / Constraint / ReviewItem`
- `Assertion / RetrievalPlan / ContextPack`

## `internal/adapters/weknora`

### `graph.go`

兼容 WeKnora 当前 `GraphData -> GraphNode/GraphRelation` JSON。解析约定：

- `__type__`：业务类型；
- `__id__`：明确实体 ID；
- `__business_key__`：跨文档业务主键；
- `__aliases__`：别名。

关系证据优先使用两端实体共享 Chunk；无共享 Chunk 时使用端点证据并标记 `derived`。

### `search.go`

只依赖 WeKnora 公开 API：

- `POST /api/v1/knowledge-search`
- `GET /api/v1/chunks/by-id/:id`

## `internal/ontology`

### `discovery.go`

自主实现的本体发现流程，不依赖 Semantica：

1. 对实体类型、属性、关系模式做确定性统计；
2. Domain/Range 支持度达到阈值才自动形成候选约束；
3. 不稳定模式进入 Review Queue；
4. 可选 `SemanticInducer` 接口用于同义词、层级和合并建议。

### `validator.go`

确定性检查 ID 完整性、Class 引用、Property Domain、Relation Domain/Range，并校验实体图是否满足正式本体。

### `compiler.go`

把 Ontology 编译成 WeKnora `ExtractConfig`。业务中文名称是权威显示词；机器 ID 不进入业务提示词。

### `registry.go`

候选与批准版本文件注册表。批准版本不可重复覆盖；高风险 Review Item 未批准时禁止发布。

## `internal/graph`

### `resolver.go`

保守实体归一：业务主键相同优先；否则仅允许同类型 + 规范化名称/别名一致。模糊相似不进入硬路径。属性冲突不覆盖。

### `assertion.go`

把实体属性和实体关系转换为独立 Assertion，形成可裁决、可版本化事实层。

## `internal/retrieval`

### `catalog.go`

把 Ontology 编译成 Fast Planner 使用的 Semantic Catalog。

### `planner.go`

高频确定性检索规划。复杂、多跳、跨源问题仅标记 `requires_llm_planning`，可通过 `ComplexPlanRefiner` 接模型。

### `arbiter.go`

按有效时间、新鲜度、来源优先级、置信度和观测时间裁决。无决定性优势时返回 `unresolved_conflict`。

## `internal/context`

`Service` 并发运行计划中的 Retriever，统一收集 Assertion/Knowledge/Path，调用 Arbiter 后由 Assembler 生成 Context Pack。

内置：
- Static Entity Graph Retriever（开发/验证）；
- Static Ontology Retriever；
- WeKnora RAG Retriever。

业务数据 Retriever 通过相同接口接入。

## `internal/mcp`

纯标准库 stdio JSON-RPC/MCP Server。运行环境由 `COBRA_*`、`WEKNORA_*` 环境变量配置。生产主入口为 `context.retrieve`。
