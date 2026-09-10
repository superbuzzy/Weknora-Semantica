# LeeClaw OpenViking UI Plugin

在 OpenClaw Control UI 中提供原生 `Memory` 与 `Skills` 页面。它只调用 OpenViking HTTP API，不修改 OpenViking Web Studio 或 Memory Engine。

运行时长期记忆仍使用 OpenViking 官方 `@openviking/openclaw-plugin` context-engine 插件；本插件只负责管理/查看 UI，因此不会复制 `assemble / afterTurn / compact` 生命周期。

Skill 的 Source of Truth 为 OpenViking：

- 用户 Skill：`viking://user/{user_id}/skills`
- 共享 Skill：`viking://agent/skills`
- Agent 可通过官方 OpenViking 插件的 `ov_search / ov_read / ov_multi_read` 按需发现和读取 Skill。
