# CobraKnowledge v0.2 架构设计

## 1. 设计目标

v0.2 将 CobraKnowledge 定义为“企业 Agent 检索与上下文治理内核”。三张图是知识资产，检索策略是核心能力。

边界：

- WeKnora：文档解析、Chunk、Wiki/RAG、现有候选实体图；只通过 Adapter/API 使用。
- Semantica：仅作为 Graph-native ontology、provenance、conflict、validation 等方法参考；不作为依赖。
- CobraKnowledge：自研 Ontology、Graph Governance、Retrieval Planner、Arbiter、Context Service、MCP。
- Agent Runtime：OpenClaw/Codex 等负责会话、模型、Skill、工具执行、权限和轨迹。

## 2. 在线架构

```text
User
  ↓
Agent Runtime
  ↓
Domain Skill                 业务方法、证据门槛、策略原则
  ↓
Context MCP                  Agent 只看到少量稳定知识工具
  ↓
Context Service
  ├─ Semantic Catalog        本体编译后的轻量语义索引
  ├─ Fast Planner            高频确定性路由
  ├─ Complex Plan Refiner    复杂任务可选 LLM，接口化
  ├─ Retrievers              Wiki/Entity/Ontology/Data
  ├─ Knowledge Arbiter       时效、来源、范围、版本、冲突
  └─ Context Assembler       统一 Context Pack
  ↓
Agent reasoning / action
```

关键原则：本体图不是每轮必查。常用本体语义编译为 Semantic Catalog，Planner 直接使用；只有 Schema/复杂合法路径问题才访问完整本体图。

## 3. 离线知识建设

```text
WeKnora Documents/Chunks
        ↓
Candidate Entity Graph
        ↓
WeKnora Normalizer
        ↓
Conservative Entity Resolver
        ↓
Pattern Analyzer
        ↓
[optional LLM Semantic Inducer]
        ↓
Candidate Ontology
        ↓
Review Queue + Deterministic Validator
        ↓
Approved Immutable Ontology
        ↓
Ontology Compiler
        ├─ Semantic Catalog -> Planner
        └─ WeKnora ExtractConfig -> constrained re-extraction
```

这形成“Bottom-up 发现 + Top-down 治理”。

## 4. 三图职责

### Wiki 图

知识页面、主题、章节、引用关系。用于解释、背景、制度、原因、原文证据和知识导航。

### 实体图

具体实体、属性、状态和实体间关系。用于事实查询、过滤、路径、诊断中的结构化事实。

### 本体图

Class、Property、Object Relation、Hierarchy、Domain/Range、Constraint、SourceBinding、RetrievalPolicy。用于语义消歧、构图约束、检索控制和复杂 Schema 查询。

## 5. 事实层：Entity 与 Assertion 分离

实体本身只保存身份与相对稳定描述；可能变化或冲突的属性值保存为 Assertion：

```text
Entity(金牛线)
  ├─ Assertion(转供状态=不可转供, source=GIS, valid=...)
  └─ Assertion(转供状态=可转供, source=PMS, valid=...)
```

Evidence 单独记录文档/Chunk/业务数据来源。更新不盲目覆盖事实，删除不盲删共享实体。

## 6. Planner 与 Arbiter

Planner 解决“怎么查”：任务类型、数据源、并发、顺序、停止条件。

Arbiter 解决“信什么”：

1. 先做时间与范围过滤；
2. 再做来源/属性策略；
3. 再参考置信度与观测时间；
4. 优势不明显时输出 unresolved_conflict，不让 LLM 强选。

LLM 只承担语义软判断，不承担硬策略。

## 7. Skill 与 MCP

Skill：业务方法、检索原则、证据门槛、冲突处理原则，不含 API/Cypher/事实。

MCP：稳定能力接口。生产面优先 `context.retrieve` 和 `context.get_evidence`；ontology/retrieval/arbitration 细粒度工具用于开发、治理和审计。

## 8. 上游兼容策略

- 不修改 WeKnora struct、worker、Neo4j repository。
- WeKnora 当前实体类型通过 `__type__=业务类型` 兼容；未来有原生 type 字段时只改 Normalizer。
- WeKnora RAG 通过公开 `/api/v1/knowledge-search`；Chunk 证据通过 `/api/v1/chunks/by-id/:id`。
- 不 import Semantica，不保存 Semantica 专有对象，不依赖其版本。

## 9. v0.2 已实现与后续

已实现：核心模型、Normalizer、Entity Resolver、Assertion/Evidence、Ontology Discovery/Validator/Compiler/Registry、Planner、Arbiter、Context Service、MCP、WeKnora RAG/Chunk Adapter、Prompt、Skill、测试。

下一阶段建议：正式 Governed Graph Store（Neo4j/PostgreSQL）、本体审核 API/UI、业务数据 SourceBinding 执行器、复杂 Planner 模型适配、增量同步/版本事件、评测集与轨迹评估。
