# leeclaw-openviking v0.8

OpenClaw 原生 Memory + Dynamic Skill Runtime。OpenViking 是 Memory 与 Skill 的唯一 Source of Truth。

## Runtime

```text
before_prompt_build
├─ Session-bound Workspace Principal
├─ Memory Recall
├─ /skills/find
├─ personal Skill > shared Skill > score
├─ Level-2 SKILL.md
├─ allowed-tools ∩ OpenClaw Tool Authority
└─ Skill Context + toolsAllow

agent_end
└─ Memory Capture / Commit

before_reset
└─ Flush pending Memory
```

## Tool Authority

Skill 的 `allowed-tools` / `allowed_tools` 只能收窄 OpenClaw 当前 Turn 已批准的工具，不能扩权。显式空列表是 deny-all optional tools。不能安全翻译的 scoped token（如 `Bash(git:*)`）不会扩大为 unrestricted `exec`。

## Workspace

`X-OpenViking-Account` 来自 Session-bound Workspace，`X-OpenViking-User` 来自 OpenClaw profileId。已有 Agent Session 不随页面默认 Workspace 漂移；新/reset Session 使用新的默认 Workspace。OpenViking Session ID 同时包含 OpenClaw session identity，reset 后不会继续写旧 Memory Session。

## 边界

v0.8 每个 Turn 自动激活一个 Skill，并加载完整 `SKILL.md` 正文；不自动执行 Skill 包中的辅助脚本/文件。
