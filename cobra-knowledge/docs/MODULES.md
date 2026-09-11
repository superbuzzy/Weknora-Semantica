# v0.8 模块与代码职责

## OpenClaw Integration

### `integrations/openclaw/workspace-core/`
唯一 Workspace membership / role / selection / downstream mapping 实现。v0.8 额外提供 Session Creator 解析和可选 Knowledge Base Scope。

### `integrations/openclaw/knowledge-plugin/`
OpenClaw 原生 Knowledge 管理与 Agent Knowledge Runtime。

关键文件：

- `index.js`：Control UI / Gateway Method / Agent Tool 注册；
- `lib/principal.js`：durable profile + Workspace → Knowledge Principal；
- `lib/weknora-client.js`：WeKnora 管理与 Graph Adapter；
- `lib/context-runtime-client.js`：LeeClaw Core Runtime 内部服务客户端；
- `lib/agent-tools.js`：`leeclaw_context_retrieve` / `leeclaw_context_get_evidence`；
- `lib/audit-log.js`：Workspace-scoped audit。

### `integrations/openclaw/openviking-plugin/`
OpenClaw 原生 Memory + Skill Runtime。

关键文件：

- `lib/client.js`：OpenViking HTTP Adapter；
- `lib/principal.js`：Session / Profile → OpenViking Principal；
- `lib/memory-runtime.js`：Memory recall / capture；
- `lib/skill-runtime.js`：Skill find / load / allowed-tools 安全映射；
- `lib/agent-runtime.js`：唯一 `before_prompt_build` 组合入口。

## LeeClaw Core

### `internal/runtimecontext/`
v0.8 新增的在线 Context Service。按请求创建 Workspace-scoped WeKnora Retriever，并复用已有 Planner / Arbiter / Assembler。

### `internal/retrieval/`
Semantic Catalog、Planner、来源策略、Arbiter。

### `internal/context/`
Retriever 编排和 Context Pack。

### `internal/ontology/`
Ontology Registry、版本、发布、active/pinned、KB Binding。

### `internal/httpapi/`
Core API：Graph、Ontology Registry、Runtime Context。

### `internal/adapters/weknora/`
WeKnora Search / Chunk / Graph Adapter。Runtime 请求通过 API Key + Tenant + External User Principal。

## Compatibility

### `compatibility/upstreams-v0.8.json`
记录 v0.8 三个 pinned upstream 与必需 Runtime Contract。

### `scripts/check-v0.8-upstreams.sh`
检查 OpenClaw Tool/Hook Authority、OpenViking Skill Contract、WeKnora Principal Contract。

### `scripts/verify-v0.8.sh`
本地发布门禁：Go、JS、Runtime、权限、不变量、旧版本活跃入口清理。
