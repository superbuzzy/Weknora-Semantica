# OpenViking integration in v0.5

v0.5 intentionally does **not** fork or vendor the OpenViking OpenClaw context-engine plugin.

Use the official upstream plugin as the Agent memory runtime. It owns `assemble`, `afterTurn`, `compact`, automatic recall/capture and the `ov_*` / memory tools. LeeClaw only adds a separate native OpenClaw Control UI plugin for Memory and Skills management.

Recommended runtime policy:

- `plugins.slots.contextEngine = openviking`
- `recallTargetTypes = ["user", "agent"]`
- keep `enableAddResourceTool = false` so enterprise knowledge stays in WeKnora
- map the active enterprise workspace to OpenViking `accountId`
- map the authenticated OpenClaw user to OpenViking `userId`

Skill source of truth is OpenViking `/api/v1/skills`. OpenClaw reads/fetches skills at runtime; local copies are caches only.
