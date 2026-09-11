# LeeClaw Knowledge Core v0.7

> 本目录继续保留历史 `cobra-knowledge` 工程路径，避免为了命名重构制造大面积无关 Git diff；对外产品与插件统一使用 LeeClaw 命名。

v0.7 的 Knowledge Core 与 OpenClaw/WeKnora/OpenViking 的边界已经固定：

```text
OpenClaw            → 人类账号 / UI / Agent Runtime
Workspace Registry  → 成员 / 角色 / 当前空间 / 下游映射
WeKnora             → Knowledge Engine
OpenViking          → Memory + Skill Engine
Knowledge Core      → Ontology / GraphView / Retrieval Governance
```

## v0.7 关键变化

- 静态 `workspaceId / tenantId / accountId` 已被服务端 Workspace Resolver 替换；
- Workspace 成员只绑定 OpenClaw durable `profileId`；
- Knowledge Plugin 增加完整 KB/文档/Wiki/FAQ/Tag/共享管理闭环；
- Memory/Skill 跟随同一 Workspace selection；
- WeKnora JWT-only KB Activity 调用已删除，替换为 LeeClaw 服务端 Workspace 审计；
- Entity Graph / Ontology Graph 继续统一使用稳定 GraphView；
- 三套 upstream 继续保持零侵入。

## 目录

```text
integrations/openclaw/knowledge-plugin   OpenClaw 原生 Knowledge/Workspace UI + WeKnora Adapter
integrations/openclaw/openviking-plugin OpenClaw 原生 Memory/Skill UI + Memory Runtime
internal/ontology                       Ontology Registry
internal/graphview                      Entity/Ontology GraphView
internal/retrieval                      Retrieval Planner / Context
configs/                                v0.7 配置示例
compatibility/                          上游契约基线
scripts/                                本地/上游升级门禁
```

## 验证

```bash
./scripts/verify-v0.7.sh

./scripts/check-v0.7-upstreams.sh \
  ../upstream/openclaw \
  ../upstream/weknora \
  ../upstream/openviking
```

## 当前生产边界

Workspace Registry、Workspace selection、LeeClaw audit、Ontology FSRegistry 当前都以单 Gateway/单实例为基线。多实例生产部署应替换为共享数据库/集中审计后端；接口 Contract 不需要变化。

Control UI 文件上传当前经过 Gateway JSON RPC 转 base64，默认 10 MiB、硬上限 50 MiB。大文件后续应采用受认证的流式上传路由，不应继续提高 JSON RPC 限额。
