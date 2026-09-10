# LeeClaw Knowledge OpenClaw Plugin

OpenClaw 原生 Knowledge 页面。UI 由 OpenClaw Control UI Plugin 承载，WeKnora 只作为知识能力后端；本体图通过现有 Graph API 读取，不复制 WeKnora 前端代码。

## 边界

- 不修改 OpenClaw core/UI 源码；
- 不修改 WeKnora 后端；
- OpenClaw 页面不直接写 WeKnora API 路径，统一经插件 runtime adapter；
- 实体图与本体图使用同一个 `GraphView`；
- `X-External-User-ID` 可由 OpenClaw `authenticatedUserId` 自动透传。

## 当前 v0.5 页面

- 知识库列表与创建；
- 文档列表；
- Wiki 页面列表；
- 实体图 / 本体图；
- Workspace 成员；
- KB 分享；
- KB 审计活动。

更复杂的文件上传、Chunk 编辑、FAQ、Datasource、邀请与分享写操作沿相同 Adapter Contract 后续扩展，不需要改 OpenClaw 主干。
