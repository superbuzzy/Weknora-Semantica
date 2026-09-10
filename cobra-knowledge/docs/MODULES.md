# v0.5 模块与代码职责

## Knowledge Core

### `internal/ontology`

本体发现、审核、发布、Registry、active/pinned KB Binding 与不可变版本。继续作为本体 Source of Truth。

### `internal/graphview`

- WeKnora Neo4j Entity Graph 只读 Adapter；
- Ontology -> GraphView；
- RegistryOntologySource。

### `internal/retrieval`

Semantic Catalog、Planner、Arbiter。负责 Agent 查询策略与可信裁决，不负责 Knowledge 管理页面。

### `internal/context`

并发 Retriever 与 Context Pack 装配。

### `internal/access`

Graph API 访问检查委托 WeKnora。v0.5 增加 API Key 与外部用户身份头透传，保持授权规则由 WeKnora 决定。

## OpenClaw Integration

### `integrations/openclaw/knowledge-plugin`

OpenClaw 原生 Knowledge Control UI + plugin runtime BFF。

浏览器：

- KB list/create；
- Documents；
- Wiki；
- Entity/Ontology Graph；
- Members/Shares；
- Activity。

Runtime Adapter：`lib/weknora-client.js`，集中保存 WeKnora REST 路径、身份头和返回值归一逻辑。

### `integrations/openclaw/openviking-plugin`

OpenClaw 原生 Memory/Skills Control UI。

Runtime Adapter：`lib/client.js`，调用 OpenViking Session/Search/Skills API。

它不承担 OpenViking context-engine；真正 Memory Runtime 使用官方插件。

### `integrations/openclaw/apply-integration.sh`

从干净 OpenClaw 生成派生构建树，并将两个自研插件加入 `extensions/*`。不修改上游源目录。

## OpenViking Integration

### `integrations/openviking/README.md`

规定官方 context-engine 的使用方式与 Memory/Skill 边界。

## Compatibility

### `compatibility/upstreams-v0.5.json`

记录当前已验证上游源码版本及依赖契约。

### `scripts/check-v0.5-upstreams.sh`

源码级 Compatibility Gate。

### `scripts/verify-v0.5.sh`

Go + JS + Browser isolation 的本地发布门禁。
