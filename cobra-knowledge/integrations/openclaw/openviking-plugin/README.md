# leeclaw-openviking v0.6

OpenClaw 原生 Memory / Skills Plugin。OpenViking 是 Memory、Session、Experience、Skill 的 Source of Truth。

## Identity

管理面唯一用户：`authenticatedUserProfile.profileId`。

```text
X-OpenViking-Account = workspaceId
X-OpenViking-User    = profileId
```

旧的静态 `accountId/userId/forwardAuthenticatedUser` 配置已删除。

## Identity-aware Memory Runtime

多用户 LeeClaw 不使用 OpenViking 官方插件静态 user/account 的 context-engine 配置。v0.6 用 OpenClaw 官方 hooks 与 Session Runtime：

- `before_prompt_build` -> 按 session creator profile 召回 Memory；
- `agent_end` -> 写入最新 user/assistant turn；
- `before_reset` -> commit pending memory；
- session owner 只能来自 `createdActor(type=human, source=profile)`。

无法解析 durable profile 时，不查询、不写入 Memory，避免跨用户污染；Agent 本身继续运行。

OpenClaw 自己的 context/compaction 机制保持不变。
