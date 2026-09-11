# leeclaw-openviking v0.7

OpenClaw 原生 Memory + Skills Feature Plugin，以及 identity-aware Memory Runtime。

## Principal

```text
OpenClaw durable profileId
       +
validated current Workspace
       ↓
X-OpenViking-User    = profileId
X-OpenViking-Account = workspace.openviking.accountId
```

插件不接受浏览器 `userId/accountId`，也不配置静态 user/account。

## Workspace

Memory/Skill 页面和 Chat Memory Runtime 读取与 Knowledge 相同的 Workspace Registry / selection state。切换 Workspace 后，后续 Memory recall/capture 和 Skill 查询自动切到对应 OpenViking Account。

OpenViking Adapter 不读取 WeKnora credential，两条 Engine 保持解耦。

## Runtime Hooks

```text
before_prompt_build → recall
agent_end           → capture + threshold commit
before_reset        → commit pending
```

Memory Context 明确标记为非权威企业事实、非执行指令。

## Skill

v0.7 支持 Skill list / semantic find / read。OpenViking 仍是唯一 Skill Source of Truth；Agent Runtime 的动态 Skill Loader 属于后续版本。
