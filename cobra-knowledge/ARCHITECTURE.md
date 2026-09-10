# CobraKnowledge v0.4 架构设计

## 1. 定位

CobraKnowledge 是企业 Agent 的检索与上下文治理内核。Wiki 图、实体图、本体图和业务数据是知识来源；核心能力是围绕这些来源执行可审计的语义解析、检索规划、冲突裁决和 Context 装配。

v0.4 将本体从“可显示的 JSON 文件”升级为“可治理的一等资产”，同时保持 WeKnora 接入面稳定。

## 2. 总体架构

```text
User
  ↓
Agent Runtime
  ↓
Domain Skill
  ↓
Context MCP
  ↓
Retrieval Control Plane
  ├─ Semantic Catalog
  ├─ Planner
  ├─ Wiki / Entity / Ontology / Data Retrievers
  ├─ Arbiter
  └─ Context Assembler
  ↓
Context Pack

                 Ontology Registry
                ┌────────┼────────┐
                ▼        ▼        ▼
             Versions  Binding   Audit
                │
          active / pinned
                │
                ▼
             Ontology
```

## 3. 三图职责

### Wiki 图

继续由 WeKnora Wiki 体系管理。主要承担主题、页面、章节和引用导航，适合解释、背景与原文证据。

### 实体图

继续由 WeKnora GraphRAG 写入 Neo4j。CobraKnowledge 只通过 Adapter 读取，不修改 WeKnora GraphRAG Schema。

### 本体图

由 CobraKnowledge 管理 Class、Property、Object Relation、Hierarchy、Domain/Range、SourceBinding、RetrievalPolicy 等语义资产。

本体同时用于：

1. 编译 Semantic Catalog；
2. 约束实体抽取与图谱校验；
3. 提供本体图可视化；
4. 管理来源、时效和检索策略语义。

## 4. v0.4 Ontology Registry

```text
Candidate Ontology
        │ approve snapshot
        ▼
Approved Ontology
        │ register
        ▼
Immutable Version Payload
        │ publish
        ▼
Published Version
        │
        ├── active_version ──────┐
        │                        │
        └── historical versions │
                                 ▼
                       KnowledgeBase Binding
                         active / pinned
```

### 不可变原则

本体版本一旦注册，`ontology_id + version` 对应的内容不允许覆盖。Registry 保存 SHA-256 并在读取时校验。

### 发布与回滚

发布、回滚只修改 Manifest 中的 release metadata 与 `active_version`，历史本体 payload 不被修改。旧 published 版本可以重新激活，实现指针级回滚。

### KB Binding

- `active`：跟随逻辑本体当前 active published version；
- `pinned`：锁定到指定 published version。

## 5. WeKnora 图谱显示

```text
WeKnora GraphExplorer
       │
实体图 / 本体图
       │ same API contract
       ▼
GET /api/v1/knowledge-bases/{kb}/graph?view=...
       │
 ┌─────┴──────────────┐
 ▼                    ▼
EntitySource       RegistryOntologySource
 ▼                    ▼
WeKnora Neo4j       Registry Resolution
                      │
                active / pinned
                      ▼
                   Ontology
```

**v0.4 没有新增 WeKnora 前端或后端改动。** v0.3 Overlay 继续工作；变化全部位于 CobraKnowledge API 内部。

## 6. 权限边界

```text
图谱读取：Browser WeKnora Token
        ↓
Cobra Graph API
        ↓ delegate
WeKnora KB RBAC

本体治理：Operator
        ↓ X-Cobra-Admin-Token
Cobra Registry API
```

图谱读取和本体治理是两套权限面。普通知识库用户不能因为能查看本体图就获得发布、回滚和绑定权限。

## 7. 在线检索

常用本体语义编译为 Semantic Catalog，在线高频问题不要求先查完整本体图。

```text
Query
  ↓
Semantic Resolver
  ↓
Fast Planner / Complex Planner
  ↓
parallel retrieval
  ├─ Wiki/RAG
  ├─ Entity Graph
  ├─ Ontology
  └─ Business Data
  ↓
Knowledge Arbiter
  ↓
Context Pack
```

## 8. 与 WeKnora / Semantica 的边界

- WeKnora：知识解析、Chunk、RAG、Wiki、GraphRAG 实体图；CobraKnowledge 通过公开接口、Neo4j 只读 Adapter、派生 Overlay 集成。
- Semantica：仅作为 graph-native ontology、provenance、conflict、validation 方法参考；运行时无依赖。
- CobraKnowledge：Ontology Registry、Graph Governance、Planner、Arbiter、Context、MCP、Graph API。

## 9. Registry 存储演进

v0.4 默认 `FSRegistry` 使用原子文件写入，适合单写实例：

```text
ontology.Registry interface
          │
     FSRegistry v0.4
          │
       future
          ▼
 PostgreSQLRegistry
```

因为上层只依赖 Registry 接口，未来切 PostgreSQL 不需要修改 WeKnora Overlay、GraphView、Planner 或 MCP 契约。

## 10. 下一阶段

v0.5 优先考虑：

1. PostgreSQL Registry，多副本事务化治理；
2. 本体候选审核 UI；
3. 本体 diff、影响分析和发布前评测门禁；
4. 发布后自动重编 Semantic Catalog / ExtractConfig；
5. 本体版本与 CobraEval 检索评测结果关联。
