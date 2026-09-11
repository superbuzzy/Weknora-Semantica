# v0.7 Knowledge 图谱：实体图 / 本体图

v0.7 保持一个稳定原则：**产品统一展示，数据生命周期不合并。**

```text
OpenClaw Knowledge
        ↓
       图谱
   ┌────┴────┐
   ▼         ▼
实体图      本体图
   │         │
WeKnora   Ontology Registry
```

## 1. 稳定 GraphView

OpenClaw Knowledge UI 只消费：

```http
GET /api/v1/knowledge-bases/{kbID}/graph?view=entity|ontology
```

统一返回节点、边和 `meta`。前端不需要知道实体图来自 Neo4j、本体图来自 Registry。

## 2. Entity Graph

Source of Truth：WeKnora GraphRAG / Neo4j。

它表示运行中的事实实体、属性和关系，随知识内容更新。

## 3. Ontology Graph

Source of Truth：LeeClaw Ontology Registry。

它表示 Class / Property / Relation / Constraint / Domain / Range 等语义治理资产，并保持：

```text
Candidate
 → immutable version
 → publish
 → active / pinned
 → KB binding
 → rollback / audit
```

## 4. 授权

Graph API 不接受或转发 LeeClaw 用户的 WeKnora Bearer。Knowledge Plugin 已经通过服务端 Workspace Resolver 得到：

```text
X-API-Key
X-Tenant-ID
X-External-User-ID = OpenClaw profileId
```

Graph API 只转发这组服务 Principal 给 WeKnora 做 KB access check。

## 5. v0.7 Workspace 影响

切换 Workspace 后，Knowledge Principal 会重新解析 WeKnora Tenant 和 service key，因此同一个 `kbID` 的访问也必须重新通过当前 Workspace 权限链，浏览器无法直接切换 Tenant 绕过校验。

## 6. 上游隔离

当前 OpenClaw 图谱页面在 LeeClaw Plugin 中实现；WeKnora upstream GraphExplorer 不需要修改。历史 overlay 只作为历史兼容资产，不应扩大 Patch 面。
