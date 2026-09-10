# LeeClaw 身份与权限模型

## 1. v0.5 当前状态

v0.5 还不是“只使用 OpenClaw 账号体系”。当前是三层联合：

- OpenClaw：用户进入统一 UI，Gateway 对插件方法执行 `operator.read` / `operator.write` 等 scope 门禁，并向插件暴露 `authenticatedUserId`；
- WeKnora：Knowledge Adapter 继续使用 WeKnora Bearer Token 或 API Key，并携带 Tenant 与 External User；知识库、成员、共享、租户隔离和资源级权限最终仍由 WeKnora 后端裁决；
- OpenViking：Memory/Skill Adapter 使用 OpenViking API Key，并携带 Account/User；Memory、Skill 的用户与 Account 隔离仍由 OpenViking 后端裁决。

因此当前属于“统一入口 + 身份透传 + 分域授权”，不是单一账号源。

## 2. 目标：OpenClaw 成为唯一账号源

目标模型：

```text
OpenClaw UserProfile
        │
        │ 唯一用户身份
        ▼
LeeClaw Principal
        │
   ┌────┴───────────────┐
   ▼                    ▼
WeKnora              OpenViking
External Principal   Trusted Principal
   │                    │
知识资源授权            Memory/Skill ACL
```

统一原则：

1. 用户只在 OpenClaw 登录和维护账号；
2. `OpenClaw UserProfile.id` 成为 LeeClaw 全局稳定 user id；
3. WeKnora/OpenViking 不再拥有独立可登录账号入口，只接收 OpenClaw 已认证用户的内部 Principal；
4. 浏览器不能自行声明 user/workspace/role，Principal 必须由 OpenClaw Gateway/Plugin 生成；
5. 后端仍保留资源级授权校验，不能只在 OpenClaw 前端隐藏按钮。

## 3. 为什么“账号统一”不等于“把所有权限判断搬进 OpenClaw”

OpenClaw 当前原生 operator role/scope 更适合平台级权限，例如：

- 能否进入 Knowledge/Memory/Skill 页面；
- 能否执行读/写类 Gateway Method；
- 能否创建 Agent Session、调用工具等。

WeKnora 已有更细的 Knowledge 资源权限：Workspace、成员、知识库共享、Viewer/Editor/Admin 等；OpenViking 则有 Account/User/ACL。直接废掉这些后端授权，只在 OpenClaw 判断一次，会造成两个问题：

- 需要在 OpenClaw 重新复制一套复杂的资源权限模型；
- 绕过 OpenClaw 直接访问后端 API 时失去最后一道授权门禁。

因此推荐架构是：

> OpenClaw 是唯一 Authentication / Account Source of Truth；WeKnora 和 OpenViking 保留 Authorization Enforcement，但不再拥有独立用户登录体系。

## 4. 统一 Principal

建议定义 LeeClaw Principal：

```json
{
  "user_id": "<openclaw-profile-id>",
  "workspace_id": "workspace-xxx",
  "platform_scopes": ["operator.read", "operator.write"],
  "issued_by": "openclaw",
  "session_id": "..."
}
```

其中：

- `user_id` 永远来自 OpenClaw 已认证 Profile；
- `workspace_id` 是当前业务空间，不允许浏览器任意覆盖；
- `platform_scopes` 来自 OpenClaw operator scope；
- Principal 由内部 Gateway/Adapter 注入并签名或仅通过可信内网头传递。

## 5. WeKnora 映射

WeKnora 不再作为用户账号源，但继续作为 Knowledge Workspace / RBAC 权威执行器。

```text
OpenClaw profile id  -> WeKnora External User ID
LeeClaw workspace id -> WeKnora Tenant/Workspace
```

OpenClaw Knowledge UI 中的“成员、共享、权限”操作仍通过 Knowledge Adapter 调 WeKnora API，从一个界面管理；WeKnora 内部只保留影子用户/外部主体和资源 ACL，不要求用户再次登录。

## 6. OpenViking 映射

OpenViking 使用 trusted principal：

```text
X-OpenViking-User    = OpenClaw profile id
X-OpenViking-Account = LeeClaw workspace id
```

用户不再单独登录 OpenViking。Memory 和个人 Skill 按 User 隔离，共享 Skill / Agent Experience 按 Account/Workspace 隔离。

## 7. 权限分层

| 层级 | 权威来源 | 负责内容 |
|---|---|---|
| 用户身份 | OpenClaw | 唯一账号、Profile、认证身份 |
| 平台权限 | OpenClaw | operator scope、插件读写能力、Agent/Tool 使用权限 |
| Knowledge 资源权限 | WeKnora | KB、Workspace、成员、共享、文档、Wiki、图谱等资源授权 |
| Memory/Skill 资源权限 | OpenViking | User/Account/ACL、个人/共享 Skill、Memory 隔离 |
| 本体治理权限 | LeeClaw Knowledge Core | Ontology publish/rollback/binding 等治理动作 |

所有权限管理页面都可以统一放在 OpenClaw UI 中，但资源级 Enforcement 仍在对应后端。

## 8. 下一步改造

从 v0.5 迁移到单账号体系建议分四步：

1. 以 OpenClaw durable `UserProfile.id` 替换插件配置中的静态 `externalUserId` / `userId`；
2. 增加可信 `LeeClaw Principal`，禁止浏览器参数覆盖当前认证用户；
3. WeKnora 进入 External Principal 模式，OpenViking 进入 trusted Account/User 模式，关闭两个系统的独立用户登录入口；
4. OpenClaw Knowledge/Memory/Skill 页面统一提供成员、共享、权限管理，后台仍调用对应 Engine 的 ACL API。

最终达到：**一个账号入口、一个用户 ID、一个产品界面；多个资源引擎各自做最后授权。**
