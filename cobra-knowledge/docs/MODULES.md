# v0.6 模块与代码职责

## OpenClaw integrations

### `integrations/openclaw/knowledge-plugin`
OpenClaw 原生 Knowledge UI + Gateway/BFF。`lib/principal.js` 只从 durable OpenClaw profile 生成用户主体；`lib/weknora-client.js` 只负责 WeKnora 协议适配。

### `integrations/openclaw/openviking-plugin`
OpenClaw 原生 Memory/Skill UI + OpenViking Adapter。

- `lib/principal.js`：Gateway profile / Session creator -> trusted principal；
- `lib/client.js`：OpenViking HTTP API；
- `lib/memory-runtime.js`：identity-aware recall/capture/commit hooks。

## Knowledge Core

### `internal/ontology`
Ontology Registry、审核快照、不可变版本、publish/activate/binding/audit。

### `internal/graphview`
Entity/Ontology -> stable GraphView projection。

### `internal/access`
Graph API 对 WeKnora 的授权委托。v0.6 只转发服务 API Key、Tenant 与 OpenClaw External Principal，不再转发人类 Bearer。

### `internal/retrieval` / `internal/arbitration` / Context
负责语义解析、检索计划、证据仲裁和 Context Pack；与 Knowledge 管理 UI 分离。

## Contracts / Compatibility

### `compatibility/upstreams-v0.6.json`
记录经过验证的三套上游版本和依赖契约。

### `scripts/check-v0.6-upstreams.sh`
检查 OpenClaw durable profile/session/hook、WeKnora API principal/API key scope、OpenViking trusted principal 等关键扩展契约。

### `scripts/verify-v0.6.sh`
本地发布门禁：Go、JS、Principal/Adapter/Runtime tests、browser isolation、legacy identity removal。
