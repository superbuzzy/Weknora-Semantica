# LeeClaw v0.6 Release

## 主题

**Single OpenClaw Identity / Durable Principal**

v0.6 将 v0.5 的“统一入口 + 多下游身份”覆盖为“OpenClaw 唯一人类身份 + 下游可信 Principal”。

## 主要变化

1. Knowledge Gateway Method 强制 `profileAccess=required`，只使用 `authenticatedUserProfile.profileId`。
2. 删除 WeKnora 用户 Bearer、静态 external user、browser tenant/user override。
3. WeKnora 只接受服务 API key + server tenant + OpenClaw external principal。
4. Graph authorization delegation 不再转发 `Authorization` / `X-External-User-Token`。
5. Memory/Skill 管理面删除静态 OpenViking user/account 和 browser override。
6. Memory Runtime 改为 OpenClaw identity-aware hooks：session creator profile -> OpenViking User。
7. 多用户 LeeClaw 不再配置 OpenViking 官方静态-user context-engine 槽位；保留 OpenClaw 自身 Context/Compaction。
8. 新增 v0.6 compatibility matrix、principal/runtime tests 与 legacy-identity grep gate。
9. 三套 upstream submodule 不修改。

## 升级注意

旧配置字段必须删除：

```text
weknoraBearerToken
externalUserId
forwardAuthenticatedUser
tenantId                 # Knowledge plugin raw field
accountId
userId
```

替换为：

```text
weknoraApiKey
weknoraTenantId
workspaceId
```

OpenViking 多用户运行时由 `leeclaw-openviking` hooks 接管 Memory recall/capture，不要同时配置静态 `openviking` context-engine，否则会重新引入共享 user/account 风险。

## 上游影响

目标门禁：

```text
OpenClaw upstream modified files: 0
WeKnora upstream modified files: 0
OpenViking upstream modified files: 0
```
