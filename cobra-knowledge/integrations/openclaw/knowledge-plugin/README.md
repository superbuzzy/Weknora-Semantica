# leeclaw-knowledge v0.6

OpenClaw 原生 Knowledge Control UI Plugin。页面属于 OpenClaw，Knowledge Engine 仍为 WeKnora，本体图由 LeeClaw Ontology Registry 提供。

## Identity

唯一用户来源：

```text
OpenClaw authenticatedUserProfile.profileId
```

插件 Gateway Method 强制 `profileAccess=required`。browser 不能传 user/tenant；服务端固定生成：

```text
X-API-Key          = weknoraApiKey
X-Tenant-ID        = weknoraTenantId
X-External-User-ID = profileId
```

旧的 `weknoraBearerToken/externalUserId/forwardAuthenticatedUser/tenantId` 配置已删除。

## UI / API boundary

browser 只调用 `leeclaw.knowledge.*`。WeKnora URL、API key、Tenant 和 External Principal 只存在于 runtime Adapter。

本体/实体图继续使用同一 GraphView：

```text
view=entity   -> WeKnora entity graph
view=ontology -> Ontology Registry
```
