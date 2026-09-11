# v0.7 模块与代码职责

## OpenClaw Integration

### `integrations/openclaw/knowledge-plugin`

OpenClaw 原生 Knowledge / Workspaces 插件。

核心子模块：

- `browser/`：只调用 `leeclaw.*` Gateway Contract；
- `lib/workspace-registry.js`：Workspace membership/role/selection/mapping；
- `lib/principal.js`：OpenClaw profile → Knowledge Principal；
- `lib/weknora-client.js`：唯一 WeKnora HTTP Adapter；
- `lib/audit-log.js`：Workspace-scoped LeeClaw operation audit；
- `index.js`：Gateway RPC、OpenClaw scope、Workspace role、审计编排。

### `integrations/openclaw/openviking-plugin`

OpenClaw 原生 Memory / Skill 插件。

- `lib/principal.js`：OpenClaw profile/session owner → OpenViking Principal；
- `lib/workspace-registry.js`：与 Knowledge 插件保持同一 Workspace Contract；
- `lib/client.js`：OpenViking HTTP Adapter；
- `lib/memory-runtime.js`：`before_prompt_build / agent_end / before_reset` Memory 生命周期。

OpenViking 侧不读取 WeKnora API Key。

## Knowledge Core

### `internal/ontology`

Ontology Registry、不可变版本、publish/activate、KB binding、rollback、audit。

### `internal/graphview`

把 WeKnora Entity Graph 与 Ontology Registry 投影成统一 `GraphView`。

### `internal/access`

Graph API 对 WeKnora 的服务 Principal 授权委托。只转发服务 API Key、Tenant 和 OpenClaw External Principal，不转发人类 Bearer。

### `internal/retrieval` / `internal/context`

Planner / retrieval / context assembly 基础能力。后续 Agent Runtime 深度接入在此扩展，不应塞进 Knowledge 管理 UI。

## Compatibility

### `compatibility/upstreams-v0.7.json`

记录 v0.7 验证过的三个 upstream 基线和关键契约。

### `scripts/check-v0.7-upstreams.sh`

检查 OpenClaw Plugin/Profile/Session、WeKnora Knowledge API/External Principal、OpenViking trusted principal/Skill/Session 契约。

### `scripts/verify-v0.7.sh`

跑 Go/JS/Plugin Contract/身份边界/浏览器泄漏/静态 Workspace 回归等本地门禁。
