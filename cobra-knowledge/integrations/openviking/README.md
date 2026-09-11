# OpenViking integration in v0.8

OpenViking 在 LeeClaw 中只承担：

```text
Memory / Session / Experience
Skill storage / semantic find / full SKILL.md read
```

身份来自 OpenClaw + Session-bound Workspace：

```text
X-OpenViking-Account = validated session workspace.openviking.accountId
X-OpenViking-User    = OpenClaw durable profileId
```

Dynamic Skill Runtime 在 OpenClaw `before_prompt_build` 中执行：

```text
/skills/find
→ score threshold
→ personal Skill 优先 / shared Skill 次之
→ Level-2 SKILL.md
→ allowed-tools ∩ OpenClaw Host Tool Authority
→ current Turn context + toolsAllow
```

OpenViking 不拥有最终工具授权权，不存企业权威事实，也不直接写 WeKnora。已有 Agent Session 固定 Workspace；页面切换只影响后续新/reset Session。
