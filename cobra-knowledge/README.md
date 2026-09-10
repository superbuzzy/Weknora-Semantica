# CobraKnowledge v0.4

> 面向企业 Agent 的检索、本体、知识图谱与上下文治理内核

CobraKnowledge 的核心目标，是让 Agent 在企业场景中能够准确理解业务语义、选择正确的数据和知识来源、处理时效与冲突，并把可追溯的 Context Pack 交给模型推理。

v0.4 的重点是 **Ontology Registry**：本体不再通过“KB -> JSON 文件路径”使用，而成为具备版本、发布、绑定、回滚和审计能力的一等资产。

## 核心架构

```text
Agent Runtime
     │
   Skill                     业务方法、证据门槛、检索原则
     │
Context MCP
     │
Retrieval Control Plane
 Resolver -> Planner -> Retrievers -> Arbiter -> Assembler
                    │
        ┌───────────┼────────────┐
        ▼           ▼            ▼
      Wiki图       实体图        本体图          Business Data
     WeKnora      Neo4j        Registry             MCP/API
```

## v0.4 本体生命周期

```text
候选本体
   │ register
   ▼
Immutable Version
   │ publish
   ▼
Published Version ────────┐
   │                      │
   ▼                      │
active_version            │
   │                      │
   ├──── active binding ──┤→ WeKnora KB
   │                      │
历史 published version ───┘ pinned binding / rollback
```

### 关键原则

- 本体内容版本注册后不可覆盖；
- 发布状态和本体 payload 分离；
- 回滚只移动 `active_version`；
- KB 可跟随 active，也可 pinned 固定版本；
- GraphView/Planner/MCP 只依赖 Registry 接口，不依赖本体文件路径；
- WeKnora 图谱页面接口不变，v0.4 不新增 WeKnora 上游改动。

## 三张图

- **Wiki 图**：知识页面、主题和引用关系，负责解释与证据导航；
- **实体图**：具体实体、属性、状态与业务关系，负责事实与关系检索；
- **本体图**：Class、Property、Relation、Hierarchy、Domain/Range、SourceBinding、RetrievalPolicy，负责业务语义和检索控制。

## WeKnora 页面

v0.3 已增加统一 `GraphExplorer`：

```text
图谱区域
┌────────────┬────────────┐
│   实体图    │   本体图    │
└────────────┴────────────┘
```

v0.4 **不修改这个前端组件**。请求仍然是：

```text
GET /api/v1/knowledge-bases/{kb_id}/graph?view=entity
GET /api/v1/knowledge-bases/{kb_id}/graph?view=ontology
```

本体图的数据解析由 CobraKnowledge 内部从文件绑定切换到 Registry。

## 快速开始

```bash
make test
make vet
make build
```

生成候选本体：

```bash
bin/cobra-knowledge bootstrap \
  -in examples/weknora-graph.json \
  -domain distribution_network \
  -out out/bootstrap
```

注册：

```bash
bin/cobra-knowledge registry-register \
  -root var/ontology-registry \
  -ontology out/bootstrap/candidate-ontology.json \
  -actor operator
```

审核完成后先生成新的 approved 快照，再注册、发布：

```bash
bin/cobra-knowledge approve-ontology \
  -ontology out/bootstrap/candidate-ontology.json \
  -version 1.0.0 \
  -reviewer reviewer \
  -out out/approved-ontology.json

bin/cobra-knowledge registry-register \
  -root var/ontology-registry \
  -ontology out/approved-ontology.json \
  -actor reviewer

bin/cobra-knowledge registry-publish \
  -root var/ontology-registry \
  -ontology-id <ontology_id> \
  -version 1.0.0 \
  -actor reviewer
```

绑定知识库：

```bash
bin/cobra-knowledge registry-bind \
  -root var/ontology-registry \
  -kb <weknora_kb_id> \
  -ontology-id <ontology_id> \
  -mode active
```

## API 启动

```bash
export COBRA_ONTOLOGY_REGISTRY_ROOT=/app/data/ontology-registry
export COBRA_REGISTRY_ADMIN_TOKEN='replace-with-random-token'
export COBRA_NEO4J_URL=http://neo4j:7474
export COBRA_NEO4J_USER=neo4j
export COBRA_NEO4J_PASSWORD='***'
export COBRA_WEKNORA_BASE_URL=http://weknora:8080
export COBRA_GRAPH_AUTH_MODE=weknora

bin/cobra-graph-api -listen :8090
```

Graph UI 的读权限继续委托 WeKnora RBAC；Registry 治理 API 使用独立管理员 Token。

## MCP

生产推荐从 Registry 解析正式本体：

```bash
export COBRA_ONTOLOGY_REGISTRY_ROOT=/app/data/ontology-registry
export COBRA_ONTOLOGY_KB_ID=<weknora_kb_id>
export WEKNORA_BASE_URL=http://weknora:8080
export WEKNORA_API_KEY=sk-xxxxx

go run ./cmd/context-mcp
```

`COBRA_ONTOLOGY_FILE` 仍可用于本地开发，但不再是生产推荐方式。

## 工程目录

```text
cobra-knowledge/
├── cmd/
│   ├── cobra/                 CLI / bootstrap / registry 管理
│   ├── context-mcp/           Agent MCP
│   └── graph-api/             Graph + Registry API
├── internal/
│   ├── ontology/              discovery / validator / compiler / registry
│   ├── graph/                 entity resolution / assertion
│   ├── graphview/             entity / ontology -> GraphView
│   ├── retrieval/             semantic catalog / planner / arbiter
│   ├── context/               retriever orchestration / Context Pack
│   ├── httpapi/               Graph API + Registry API
│   └── access/                WeKnora RBAC delegation
├── integrations/weknora/      非侵入 Overlay
├── prompts/
├── skills/
└── docs/
```

## 上游隔离

WeKnora 与 Semantica 都不是 CobraKnowledge 内核源码的一部分：

- WeKnora：知识/RAG/实体图底座之一，通过 API、Neo4j 只读适配器和派生 Overlay 集成；
- Semantica：本体构建、Provenance、Conflict、Validation 等方法参考，无运行时依赖；
- CobraKnowledge：独立维护本体、检索策略、冲突裁决和 Context 能力。

详见：

- `ARCHITECTURE.md`
- `docs/ONTOLOGY_REGISTRY.md`
- `docs/GRAPH_VISUALIZATION.md`
- `docs/MODULES.md`
- `RELEASE-v0.4.md`
