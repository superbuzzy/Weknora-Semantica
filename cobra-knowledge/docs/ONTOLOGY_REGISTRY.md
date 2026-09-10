# v0.4 Ontology Registry 设计与使用

## 1. 目标

v0.3 已经可以把本体图显示到 WeKnora 页面，但知识库与本体之间仍通过 `ontology-bindings.json -> ontology.json` 的文件路径关系连接。v0.4 将本体升级为 CobraKnowledge 的一等治理资产：

- 本体按逻辑 `ontology_id` 管理，不向调用方暴露文件路径；
- 每个版本是不可变快照；
- 发布状态与本体内容分离；
- 一个逻辑本体有一个 `active_version`；
- KB 可以跟随 active 版本，也可以固定到某个 published 版本；
- 回滚只移动 active 指针，不修改历史版本；
- 注册、发布、切换、绑定均记录审计事件；
- Graph API、Planner、MCP 通过 Registry 接口读取本体，不依赖具体存储实现。

## 2. 数据结构

```text
Ontology Registry
├── OntologyManifest
│   ├── ontology_id
│   ├── domain
│   ├── active_version
│   └── versions[]
│       ├── version
│       ├── state = candidate | published
│       ├── content_sha256
│       ├── created_by / created_at
│       └── published_by / published_at
│
├── OntologyVersion Payload     # immutable
│   └── Ontology JSON
│
├── KnowledgeBase Binding
│   ├── kb_id
│   ├── ontology_id
│   ├── mode = active | pinned
│   └── version (pinned only)
│
└── Audit Event                 # append only
```

## 3. 为什么内容和发布状态分开

历史本体版本必须可复现。发布、回滚或再次激活旧版本时，不应修改原来的本体内容。因此 v0.4 将：

- `versions/<version>.json` 作为不可变内容快照；
- `manifest.json` 保存发布状态和 active 指针；
- 文件注册后再次写入相同 `ontology_id + version` 会直接拒绝；
- 读取版本时重新计算 SHA-256，发现内容被外部修改则拒绝加载。

## 4. Active 与 Pinned 两种绑定

### active

```text
KB-A -> DistributionOntology -> active_version
```

发布 1.1.0 后，KB-A 自动读取 1.1.0。适用于希望统一跟随正式版本的知识库。

### pinned

```text
KB-B -> DistributionOntology@1.0.0
```

即使 1.1.0 已发布，KB-B 仍使用 1.0.0。适用于生产灰度、专项评测、重要业务冻结窗口。

## 5. 回滚

回滚不是覆盖新版本，也不是复制旧文件：

```text
active_version = 1.1.0
        ↓ rollback
active_version = 1.0.0
```

只有已经发布过的版本才能被重新激活。

## 6. 文件系统实现

v0.4 默认实现 `FSRegistry`：

```text
var/ontology-registry/
├── ontologies/
│   └── <ontology_id>/
│       ├── manifest.json
│       └── versions/
│           ├── 1.0.0.json
│           └── 1.1.0.json
├── bindings/
│   └── <kb_id>.json
└── audit/
    └── events.jsonl
```

该实现适用于单实例或共享文件系统部署。接口已经抽象为 `ontology.Registry`，后续接 PostgreSQL 时无需修改 WeKnora Overlay、GraphView、Planner 或 Agent MCP 契约。

## 7. CLI

注册不可变版本：

```bash
bin/cobra-knowledge registry-register \
  -root var/ontology-registry \
  -ontology out/bootstrap/candidate-ontology.json \
  -actor lichongyang \
  -notes '配网本体首版'
```

审核候选并生成新的 approved 快照：

```bash
bin/cobra-knowledge approve-ontology \
  -ontology out/bootstrap/candidate-ontology.json \
  -version 1.0.0 \
  -reviewer reviewer \
  -out out/approved-ontology.json
```

将 approved 快照注册为不可变版本：

```bash
bin/cobra-knowledge registry-register \
  -root var/ontology-registry \
  -ontology out/approved-ontology.json \
  -actor reviewer
```

发布并设为 active：

```bash
bin/cobra-knowledge registry-publish \
  -root var/ontology-registry \
  -ontology-id <ontology_id> \
  -version 1.0.0 \
  -actor reviewer
```

绑定 KB 跟随 active：

```bash
bin/cobra-knowledge registry-bind \
  -root var/ontology-registry \
  -kb <weknora_kb_id> \
  -ontology-id <ontology_id> \
  -mode active \
  -actor operator
```

固定版本：

```bash
bin/cobra-knowledge registry-bind \
  -root var/ontology-registry \
  -kb <weknora_kb_id> \
  -ontology-id <ontology_id> \
  -mode pinned \
  -version 1.0.0
```

回滚 active 指针：

```bash
bin/cobra-knowledge registry-activate \
  -root var/ontology-registry \
  -ontology-id <ontology_id> \
  -version 1.0.0 \
  -actor operator
```

## 8. Registry REST API

治理 API 与 WeKnora 图谱读取 API 分离授权。所有 `/api/v1/registry/*` 路由都要求：

```text
X-Cobra-Admin-Token: <COBRA_REGISTRY_ADMIN_TOKEN>
```

普通 WeKnora `Authorization` Token 不能直接获得本体发布权限。

主要接口：

```text
GET  /api/v1/registry/ontologies
GET  /api/v1/registry/ontologies/{ontology_id}
GET  /api/v1/registry/ontologies/{ontology_id}/versions/{version}
POST /api/v1/registry/ontologies/versions
POST /api/v1/registry/ontologies/{ontology_id}/versions/{version}/publish
POST /api/v1/registry/ontologies/{ontology_id}/activate
GET  /api/v1/registry/knowledge-bases/{kb_id}/binding
PUT  /api/v1/registry/knowledge-bases/{kb_id}/binding
GET  /api/v1/registry/audit?limit=100
```

## 9. 对 WeKnora 的影响

v0.4 不修改 WeKnora 后端，不修改 GraphRAG，不修改 Neo4j Schema，也不增加 WeKnora 数据表。

WeKnora Overlay 仍调用：

```text
GET /api/v1/knowledge-bases/{kb_id}/graph?view=ontology
```

变化只发生在 CobraKnowledge 内部：

```text
v0.3: KB -> binding file -> ontology file
v0.4: KB -> Registry binding -> active/pinned published version
```

因此 WeKnora UI 契约不变，未来上游升级仍只需要维护现有的小型 Overlay 接入点。

## 10. 当前边界

FSRegistry 使用进程内锁 + 原子 rename，适合单写实例。生产多副本写入场景下一阶段应实现 PostgreSQL Registry，利用数据库事务、唯一约束和行锁替代文件系统协调。读取接口和上层代码不需要变化。
