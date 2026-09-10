# CobraKnowledge / LeeClaw Integration v0.5.0

## Release theme

v0.5 将原有 Knowledge Core 放入新的组合架构：OpenClaw 作为 Agent 主干和统一 UI，WeKnora 作为 Knowledge Engine，OpenViking 作为 Memory + Skill Engine。

## Added

- `leeclaw-knowledge` OpenClaw 原生 Control UI Plugin；
- WeKnora runtime Adapter/BFF；
- KB list/create、Document、Wiki、Entity/Ontology Graph、Member/Share/Audit 原生页面；
- `leeclaw-openviking` OpenClaw 原生 Memory/Skills Plugin；
- OpenViking Session/Memory/Skills Adapter；
- OpenViking Skill Source of Truth 约束；
- OpenClaw derived build assembly script；
- 三上游源码 Compatibility Gate；
- Adapter contract tests；
- Browser 禁止直连上游 API 的发布门禁；
- WeKnora Graph API 权限委托支持 API Key 与外部用户身份头。

## Preserved

- v0.4 Ontology Registry / publish / rollback / KB binding；
- Entity Graph / Ontology Graph `GraphView` 契约；
- Context MCP / Planner / Arbiter；
- v0.3 WeKnora GraphExplorer overlay，未扩大 patch 面。

## Upstream impact

v0.5 的设计目标是：

```text
OpenClaw upstream modified files: 0
WeKnora upstream modified files: 0
OpenViking upstream modified files: 0
```

组合通过外部 Plugin、Adapter 和派生构建树完成。

## Validated source snapshots

- OpenClaw `2026.9.3`
- WeKnora `0.8.0`
- OpenViking official OpenClaw plugin `2026.6.18`

## Verification

发布前必须执行：

```bash
make verify
./scripts/check-v0.5-upstreams.sh <openclaw> <weknora> <openviking>
```

完整 OpenClaw bundle 还要求可用的 pnpm/npm 依赖环境；离线环境不能用源码检查替代真实 bundle 构建。
