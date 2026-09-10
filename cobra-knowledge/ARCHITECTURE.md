# CobraKnowledge v0.3 架构设计

## 1. 定位

CobraKnowledge 是“企业 Agent 检索与上下文治理内核”。三张图是知识资产，检索策略是核心能力。v0.3 新增一个重要原则：**本体图既是检索语义模型，也是一张可直接浏览的图，但显示契约与底层存储解耦。**

边界：

- **WeKnora**：文档解析、Chunk、Wiki/RAG、GraphRAG 实体图；不修改业务内核。
- **Semantica**：Graph-native ontology、provenance、conflict、validation 等方法参考；无运行时依赖。
- **CobraKnowledge**：Ontology、Graph Governance、Retrieval Planner、Arbiter、Context Service、MCP、Graph API。
- **Agent Runtime**：OpenClaw/Codex 等负责会话、模型、Skill、工具执行、权限和轨迹。

## 2. 在线 Agent 架构

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
  ├─ Complex Plan Refiner    复杂任务可选 LLM
  ├─ Retrievers              Wiki/Entity/Ontology/Data
  ├─ Knowledge Arbiter       时效、来源、范围、版本、冲突
  └─ Context Assembler       统一 Context Pack
  ↓
Agent reasoning / action
```

本体图不是每轮必查。常用本体语义编译为 Semantic Catalog；只有 Schema、复杂合法路径或本体治理问题才访问完整本体。

## 3. 三图体系

### Wiki 图

知识页面、主题、章节和引用关系。继续使用 WeKnora Wiki Browser，重点支持知识导航、解释、背景和原文证据。

### 实体图

具体实体、属性、状态和关系。实体图由 WeKnora GraphRAG 写入 Neo4j，CobraKnowledge v0.3 仅只读用于统一检索和可视化，不复制、不改 Schema。

### 本体图

Class、Property、Object Relation、Hierarchy、Domain/Range、Constraint、SourceBinding、RetrievalPolicy。它既：

1. 编译成 Semantic Catalog 服务 Planner；
2. 约束后续实体图抽取；
3. 通过 Graph API 投影为可视化图。

## 4. 本体图显示架构

```text
                  WeKnora Graph Settings
                           │
                   GraphExplorer.vue
                           │
                实体图 / 本体图 switch
                           │
                           ▼
                 Cobra Graph API
          GET .../graph?view=entity|ontology
                     │             │
          ┌──────────┘             └──────────┐
          ▼                                   ▼
 Neo4jHTTPSource                       OntologySource
 只读 WeKnora Neo4j                 KB -> Ontology binding
          │                                   │
          └───────────┐           ┌───────────┘
                      ▼           ▼
                    GraphView DTO
                nodes / edges / meta
```

前端不知道 Neo4j 节点结构，也不知道 Ontology JSON 结构。两种图都只认 `GraphView`。以后任何一侧更换存储实现，只换 Source Adapter。

## 5. 为什么不把本体图直接塞进 WeKnora Neo4j Schema

v0.3 的目标是“显示和使用本体图”，不是把两套生命周期绑死。

- WeKnora 实体图随文档增删重建；
- CobraKnowledge 本体需要审核、版本、批准和不可变发布；
- 两者更新节奏不同；
- 直接共用 WeKnora Label/Relation Schema 会让官方升级与本体治理相互影响。

因此 v0.3 可以共用 Neo4j 基础设施，但不共用 WeKnora GraphRAG 数据模型。本体当前仍以 CobraKnowledge Ontology Registry 为权威源，显示时动态投影为图。

## 6. GraphView 稳定契约

```text
GraphView
├─ nodes[]
│  ├─ id
│  ├─ label
│  ├─ kind        entity/class/property
│  ├─ group
│  └─ metadata
├─ edges[]
│  ├─ source
│  ├─ target
│  ├─ label
│  ├─ kind
│  └─ metadata
└─ meta
   ├─ view        entity/ontology
   ├─ knowledge_base_id
   ├─ total_nodes
   ├─ returned_nodes
   ├─ returned_edges
   ├─ truncated
   └─ ontology_version
```

本体图投影：

```text
OntologyClass        -> Class Node
DataProperty         -> Property Node
ParentIDs            -> subclass_of edge
DomainIDs            -> has_property edge
ObjectRelation       -> Domain Class --relation--> Range Class
```

## 7. 权限模型

图谱显示不能绕过 WeKnora RBAC。

```text
Browser
  │ Authorization / X-Tenant-ID
  ▼
Cobra Graph API
  │
  ├─ forward identity headers
  ▼
WeKnora GET /api/v1/knowledge-bases/{kb_id}
  │
  ├─ 2xx -> permit graph read
  └─ non-2xx -> deny
```

生产默认 `COBRA_GRAPH_AUTH_MODE=weknora`。开发环境必须显式配置 `off` 才能关闭授权。

## 8. WeKnora 升级隔离

不在 `upstream/weknora` 目录直接改源码：

```text
upstream/weknora             官方可直接 git pull
       │
       │ copy
       ▼
build/weknora-v0.3
       │
       ├─ overlay new files
       └─ apply one small GraphSettings patch
```

如果官方调整 `GraphSettings.vue` 导致补丁失配，构建脚本 dry-run 直接失败。修改范围被压缩为一个小适配点，不会出现大 fork 长期冲突。

## 9. 离线知识建设

```text
WeKnora Documents/Chunks
        ↓
Candidate Entity Graph
        ↓
Normalizer + Conservative Entity Resolver
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
        ├─ WeKnora ExtractConfig -> constrained re-extraction
        └─ GraphView -> UI
```

形成“Bottom-up 发现 + Top-down 治理 + 可视化审核”。

## 10. 事实层与冲突

Entity 只保存身份；可能变化或冲突的事实保存为 Assertion，并通过 Evidence 追溯来源。Arbiter 按有效时间、来源优先级、置信度和观测时间处理；无决定性优势时保留 `unresolved_conflict`，不让模型强选。

## 11. v0.3 已实现

- `GraphView` DTO；
- Ontology -> GraphView 投影；
- WeKnora Neo4j HTTP 只读图源；
- KB -> Ontology 文件绑定；
- `cobra-graph-api`；
- WeKnora RBAC 委托校验；
- WeKnora `GraphExplorer` 实体图/本体图切换；
- 非侵入 overlay 构建机制；
- 单测与补丁应用验证。

## 12. 后续重点

下一阶段不急于继续扩 UI，优先做：

1. Governed Graph Store 与正式 Ontology Registry API；
2. 实体图增量同步和 Assertion/Evidence 生命周期；
3. 图谱局部搜索、Ego 展开，避免大图全量渲染；
4. 本体审核界面：候选 Class/Property/Relation 的批准、驳回和版本发布；
5. 将本体版本与检索评测结果关联，形成“本体变更 -> 检索复测”。
