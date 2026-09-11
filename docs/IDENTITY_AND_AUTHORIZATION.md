# LeeClaw v0.8 身份、Workspace 与授权模型

## 1. 唯一人类账号

LeeClaw 只认 OpenClaw durable User Profile：

```text
authenticatedUserProfile.profileId
              ↓
        global user id
```

WeKnora 和 OpenViking 不建立第二套人类登录。

## 2. Workspace 是资源授权域

Workspace Registry 保存：

```text
workspace
├─ members: profileId + role
├─ WeKnora tenant + service-key env + optional KB allowlist
└─ OpenViking account
```

角色：

```text
viewer < editor < admin < owner
```

一次管理操作需要：

```text
OpenClaw platform scope
AND Workspace role
AND downstream service capability
```

## 3. v0.8：默认 Workspace 与 Session Workspace 分离

v0.7 用 `profileId -> selected workspace` 同时驱动页面和 Chat Runtime。这个逻辑在用户切换 Workspace 后可能让已有聊天历史突然进入另一资源域。

v0.8 直接替换为两层状态：

```text
Profile Default Workspace
  profileId -> workspaceId

Session Workspace Binding
  hashed session identity -> profileId + workspaceId
```

页面切换只改变 Profile Default Workspace。Agent Session 第一次运行时绑定当时的默认 Workspace，之后保持固定。

```text
Session A: 重庆公司
用户切默认 Workspace → 总部
Session A: 仍为重庆公司
/new 或 /reset
Session B: 总部
```

Session binding 优先使用 OpenClaw `sessionId`，缺失时才退到 `sessionKey`。原始 Session ID 不写入状态文件，只保存哈希键。

如果用户被移出已绑定 Workspace，既有 Session fail closed，不会自动漂移到另一个 Workspace。

## 4. Knowledge Principal

服务端生成：

```text
userId              = profileId
workspaceId         = Session/Request 已验证 Workspace
workspaceRole       = Workspace role
X-Tenant-ID         = workspace.weknora.tenantId
X-API-Key           = env[workspace.weknora.apiKeyEnv]
X-External-User-ID  = profileId
knowledgeBaseIds    = workspace.weknora.knowledgeBaseIds
```

浏览器和模型都不能声明这些可信字段。

## 5. OpenViking Principal

```text
X-OpenViking-Account = workspace.openviking.accountId
X-OpenViking-User    = profileId
```

Memory、Skill 搜索和 Skill 加载都使用同一个 Session-bound Workspace。

## 6. Agent Knowledge Tool

模型只提交业务参数：

```text
retrieve:
  query
  knowledge_base_id?
  domain?
  task?

evidence:
  chunk_id
```

身份、Workspace、Tenant、Credential、KB allowlist 全部由服务端恢复。

Evidence Tool 读取 chunk 后，再用返回的 `knowledge_base_id` 与 Workspace allowlist 比对。已知 chunk id 不能绕过 KB Scope。

## 7. Skill Tool Policy

OpenViking Skill 的 `allowed-tools` 不是授权源。

```text
Effective Turn Tools
=
OpenClaw Host-approved Tools
∩
Skill allowed-tools
```

Skill 不能扩大平台权限；显式空列表表示 deny-all optional tools。

## 8. 状态存储

v0.8 Workspace state 使用 canonical version 2：

```json
{
  "version": 2,
  "profiles": {
    "profile-1": "workspace-cq"
  },
  "sessions": {
    "<sha256>": {
      "profileId": "profile-1",
      "workspaceId": "workspace-cq",
      "boundAt": "..."
    }
  }
}
```

v0.7 flat state 会在首次写入 v0.8 state 时单向升级到该结构，不再作为第二套运行模型存在。

## 9. 生产扩展

当前 Registry / Session binding / Audit 是单 Gateway 文件后端。HA 时应把相同 Contract 换成共享数据库，并补：

- 事务与乐观锁；
- Session binding TTL / 清理；
- Workspace membership cache invalidation；
- Secret Manager；
- 集中审计与安全告警。

这些变化不应修改 Browser 参数、Agent Tool Schema 或下游 Principal Contract。
