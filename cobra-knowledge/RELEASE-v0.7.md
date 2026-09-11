# LeeClaw v0.7 Release

## 主题

**Multi-Workspace / Complete Knowledge Management**

v0.7 把 v0.6 的静态单 Workspace 直接替换为 OpenClaw-profile 驱动的 Workspace Resolver，并把 WeKnora 的主要 Knowledge 管理闭环原生集成到 OpenClaw。

## 主要变化

1. 新增 Workspace Registry：`profileId -> membership/role -> downstream mapping`。
2. 新增 Workspace selector 和成员管理页面；owner 直接管理 OpenClaw Profile membership。
3. 删除插件静态 `workspaceId / weknoraTenantId / weknoraApiKey` 活跃配置。
4. Workspace 切换同步影响 Knowledge、OpenViking Memory/Skill 和 Chat Memory Runtime。
5. Knowledge 新增 KB update/delete、文档 file/url/manual、reparse/cancel/delete、Wiki、FAQ、Tag、hybrid-search、organization/share 写操作。
6. WeKnora API 路径继续集中在 Adapter，browser 不认识任何下游 credential/trusted header。
7. 修正 WeKnora JWT-only KB activity：删除不可达调用，改用 LeeClaw 服务端 Workspace-scoped audit。
8. OpenViking Workspace 解析不再依赖 WeKnora API Key。
9. Entity/Ontology GraphView 和 Ontology Registry 生命周期保持不变。
10. 三个 upstream submodule 不修改。

## 权限矩阵

```text
viewer  -> read
editor  -> Knowledge content write
admin   -> destructive KB/share operations
owner   -> Workspace member governance
```

每次请求同时受 OpenClaw `operator.*` scope、Workspace role 和下游 service capability 约束。

## 验证边界

v0.7 的 Workspace Registry/selection/audit 为单 Gateway 文件实现；HA 需要数据库后端。Control UI 文件上传默认 10 MiB，硬上限 50 MiB；大文件需要后续流式上传路由。

完整 OpenClaw pnpm bundle 只有在依赖完整的 CI 环境实际跑通后才可标记通过，本 Release 不以语法/合同测试冒充完整 bundle E2E。
