# LeeClaw Knowledge Core v0.6

> 本目录仍保留历史 `cobra-knowledge` 工程名，避免为了命名重构制造无关 Git 噪声。v0.6 的重点是身份与运行边界收口。

## 核心定位

```text
OpenClaw       = 唯一用户身份 + Agent Runtime + UI Host
WeKnora        = Knowledge Engine
OpenViking     = Memory + Skill Engine
Knowledge Core = Ontology + Retrieval + Arbiter + Context
```

### v0.6 新的 Principal 链

```text
OpenClaw authenticatedUserProfile.profileId
        │
        ├─ Knowledge Plugin
        │      └─ X-External-User-ID -> WeKnora
        │
        └─ OpenViking Plugin
               └─ X-OpenViking-User -> OpenViking
```

`tenant/account` 均由服务端 Workspace 配置产生；browser 无法覆盖。

## 目录

```text
integrations/openclaw/
├─ knowledge-plugin/
│  ├─ browser/          # 只调 Gateway Contract
│  ├─ lib/principal.js  # OpenClaw profile -> Knowledge Principal
│  └─ lib/weknora-client.js
└─ openviking-plugin/
   ├─ browser/
   ├─ lib/principal.js  # OpenClaw profile/session owner -> OV Principal
   ├─ lib/client.js
   └─ lib/memory-runtime.js

internal/
├─ ontology/
├─ graphview/
├─ access/
├─ httpapi/
├─ retrieval/
└─ ...
```

## Memory Runtime

v0.6 不再建议在多用户 LeeClaw 中配置 OpenViking 官方插件的静态 `userId/accountId`。`leeclaw-openviking` 使用 OpenClaw 官方 hooks + Session Runtime：

- `before_prompt_build`：按 session creator profile 召回 Memory；
- `agent_end`：按同一 profile 捕获最新 turn；
- `before_reset`：提交 pending session；
- OpenClaw 原 Context/Compaction 不被替换。

## Knowledge 安全边界

WeKnora Adapter 只发送：

```text
X-API-Key
X-Tenant-ID
X-External-User-ID=<OpenClaw profileId>
```

用户 Bearer 和客户端声明的 external user 已从 v0.6 路径移除。WeKnora service API key 仍应按 capability / `knowledge_base_ids` 最小授权，并启用 API Principal `direct_header + require_direct_header`。

## 本体

v0.4 起的 Ontology Registry 继续保留：immutable version、publish、active/pinned、rollback、audit、KB binding。OpenClaw Knowledge 图谱页统一查看 Entity Graph / Ontology Graph，底层仍不混库。

## 验证

```bash
./scripts/verify-v0.6.sh
```

上游契约：

```bash
./scripts/check-v0.6-upstreams.sh <openclaw> <weknora> <openviking>
```
