# v0.8 Knowledge 图谱：实体图 / 本体图

v0.8 保持一个稳定原则：**产品统一展示，数据生命周期不合并。**

## 1. 三类知识结构

```text
Wiki / RAG
  → WeKnora

Entity Graph
  → WeKnora GraphRAG / Neo4j

Ontology Graph
  → LeeClaw Ontology Registry
```

## 2. 统一展示契约

实体图和本体图继续统一投影为 `GraphView`：

```text
nodes[]
edges[]
meta.view = entity | ontology
```

OpenClaw Knowledge 页面只依赖稳定 GraphView，不依赖 Neo4j 或 Registry 的内部结构。

## 3. v0.8 Core API

v0.8 已直接用 `coreApiBaseUrl` 替换旧的 `graphApiBaseUrl`。图谱接口仍保持：

```text
GET /api/v1/knowledge-bases/{kbID}/graph?view=entity
GET /api/v1/knowledge-bases/{kbID}/graph?view=ontology
```

同一个 Core API 同时新增：

```text
POST /api/v1/runtime/context/retrieve
POST /api/v1/runtime/context/evidence
```

因此 Graph 是 Knowledge Core 的一个视图能力，不再被单独抽象成一个产品服务。

## 4. 本体图生命周期

本体仍按：

```text
Candidate
  → Immutable Version
  → Publish
  → Active / Pinned
  → KB Binding
  → Rollback / Audit
```

本体图不会因为 UI 与实体图放在同一页面就与 WeKnora Entity Graph 合库。

## 5. Runtime 关系

Ontology Graph 在 v0.8 不只用于展示，还会参与：

```text
KB Binding
  → Ontology
  → Semantic Catalog
  → Retrieval Planner
```

它用于决定“问题涉及什么语义、优先查询什么来源”；具体企业事实仍由 WeKnora / Entity / Business Retriever 提供。

## 6. Workspace

图谱访问与 Knowledge Runtime 都服从当前 Workspace。浏览器不能指定 WeKnora Tenant 或 Service Credential；这些由服务端 Workspace Principal 生成。
