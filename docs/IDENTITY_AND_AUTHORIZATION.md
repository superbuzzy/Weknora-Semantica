# LeeClaw v0.6 身份与权限模型

## 当前状态

v0.6 已完成单一人类账号源收口：**用户只由 OpenClaw 认证，唯一持久用户 ID 为 `authenticatedUserProfile.profileId`。**

```text
OpenClaw Profile
      │
      ├─> WeKnora External Principal
      └─> OpenViking Trusted Principal
```

WeKnora/OpenViking 不再要求用户单独登录；其 API Key/root key 属于服务间凭证，不代表人类账号。

## 权限分层

| 层级 | 权威 | 说明 |
|---|---|---|
| 人类账号/登录 | OpenClaw | 唯一账号源 |
| 平台读写 scope | OpenClaw | `operator.read/write/...` |
| Knowledge 服务边界 | WeKnora | Tenant + API key capability + KB scope |
| Memory/Skill 隔离 | OpenViking | Account + User + ACL |
| Ontology 治理 | Knowledge Core | publish/rollback/binding/admin token |

## 不能由客户端声明的字段

以下均为 server-owned：

```text
userId
externalUserId
tenantId
accountId
```

OpenClaw browser 只发业务参数。身份由 Gateway Client / Session Store 解析，作用域由 plugin config/未来 Workspace Resolver 注入。

## Knowledge

```text
profileId -> X-External-User-ID
configured tenant -> X-Tenant-ID
service key -> X-API-Key
```

不转发人类 WeKnora Bearer。WeKnora 应启用 external principal direct-header 且 `require_direct_header=true`，并让服务 key 只拥有 LeeClaw 所需能力。

注意：v0.6 的 WeKnora 服务授权粒度主要由 API key capability / `knowledge_base_ids` 控制；OpenClaw operator scope 控制平台级读写。用户身份会被 WeKnora作为 external principal记录/隔离，但 v0.6 没有伪造一套 WeKnora Web User 登录会话。

## Memory / Skill

管理面：Gateway 的 durable profile 直接映射 `X-OpenViking-User`。

运行面：从 OpenClaw Session 的 `createdActor(source=profile)` 恢复 owner profile。这样后台 Agent turn 不需要信任浏览器传来的 user id。

```text
OpenViking Account = workspaceId
OpenViking User    = profileId
```

## 多 Workspace

v0.6 先固定 workspace/tenant 映射。未来允许切 Workspace 时，必须增加 server-side resolver + membership check，浏览器只能请求逻辑 workspace，不能直接提交 WeKnora tenant 或 OpenViking account ID。
