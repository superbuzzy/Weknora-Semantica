# LeeClaw v0.7 身份、Workspace 与授权模型

## 1. 唯一人类账号

LeeClaw 只维护一套人类账号：OpenClaw durable User Profile。

```text
OpenClaw authenticatedUserProfile.profileId
                  ↓
            global user id
```

WeKnora/OpenViking 不要求用户再次登录。

## 2. v0.7 对 v0.6 的修正

v0.6 已解决“用户身份不能由浏览器覆盖”，但 Workspace 仍是静态服务配置。v0.7 直接删除静态 Workspace/Tenant/Account 路径，替换为服务端 Workspace Registry。

另外，WeKnora API Principal 的 External User 适合标识调用主体，但共享 service key 下它不等价于 OpenClaw 多用户 RBAC。v0.7 因此不再把 WeKnora 用户/成员体系当 LeeClaw 人类账号体系。

## 3. Workspace 是 LeeClaw 的授权域

Workspace Registry 保存：

```text
logical workspace
├─ members: OpenClaw profileId + role
├─ WeKnora tenant mapping + API key env reference
└─ OpenViking account mapping
```

浏览器只可以提交逻辑 `workspaceId` 作为切换请求。服务端先验证当前 profile membership；未经验证的 logical id 永远不能成为下游 Tenant/Account。

## 4. 权限交集

一次写操作最终需要：

```text
OpenClaw platform scope
AND Workspace role
AND downstream machine-principal capability
```

典型矩阵：

| 操作 | OpenClaw Scope | Workspace Role |
|---|---|---|
| 查看 KB/Memory/Skill | operator.read | viewer+ |
| 新建/修改 KB 内容 | operator.write | editor+ |
| 删除 KB / 管理共享 | operator.admin | admin+ |
| 增删 Workspace 成员 | operator.admin | owner |

## 5. WeKnora Principal

服务端根据已验证 Workspace 生成：

```text
X-API-Key          = env[workspace.weknora.apiKeyEnv]
X-Tenant-ID        = workspace.weknora.tenantId
X-External-User-ID = OpenClaw profileId
```

用户 Bearer、浏览器 external user、浏览器 tenant override 都不属于 v0.7 活跃路径。

## 6. OpenViking Principal

```text
X-OpenViking-Account = workspace.openviking.accountId
X-OpenViking-User    = OpenClaw profileId
```

OpenViking 解析 Workspace 不需要 WeKnora credential；Knowledge 和 Memory/Skill 两条引擎边界独立。

## 7. Workspace 成员治理

成员对象只保存：

```json
{
  "profileId": "openclaw-profile-id",
  "role": "editor",
  "displayName": "optional cache"
}
```

成员页面从 OpenClaw `users.list` 获取 Profile 信息。owner 可以在拥有 `operator.admin` 的情况下增加、修改或移除成员；系统阻止删除/降级最后一个 owner。

## 8. Current Workspace

当前选择按 `profileId -> workspaceId` 服务端持久化。Knowledge 页面、Memory/Skill 页面和 Chat Memory Runtime 都读取同一 selection state。

因此：

```text
切换 Workspace
  → Knowledge Tenant 变更
  → OpenViking Account 变更
  → 后续 Chat Memory scope 变更
```

不会出现页面已切换、Agent 仍读旧空间 Memory 的分叉。

## 9. 生产扩展

v0.7 Registry/selection 使用单 Gateway 文件存储。HA 部署应实现相同 Contract 的数据库后端，并增加：

- 乐观锁/事务；
- Workspace 变更审计；
- 缓存失效；
- Secret Manager 引用；
- 集中授权事件与告警。

升级时替换存储实现，不改变 browser 参数或下游 principal contract。
