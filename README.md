# LeeClaw v0.10

> **OpenClaw Agent Runtime + WeKnora Knowledge + OpenViking Memory / Experience / Skill**

v0.10 的主题是 **Experience Promotion**：v0.9 已解决“该查哪里、结果是否冲突、是否过期、是否完整”，v0.10 进一步把 Memory / Experience 形成的经验纳入受控治理。经验可以成为候选 Knowledge 或 Skill，但永远不能绕过 Promotion Gate 直接写入正式资产。

## 架构边界

```text
OpenClaw   = 唯一人类账号 + UI + Agent Runtime + Tool Authority
WeKnora    = 企业 Knowledge Engine + 正式 Knowledge
OpenViking = Memory + Session + Experience + 正式 Skill
LeeClaw    = Workspace + Ontology + Retrieval + Promotion + Governance
```

## v0.10 Promotion 主链

```text
Session / Experience / Evidence
  → Candidate（默认 personal）
  → Source Trace Verification
  → Dedup / Conflict / Validation
  → checked / blocked
  → Admin Review
  → personal → workspace 显式提升
  → Promotion Gate
  → Pre-publish Recheck
  → WeKnora Knowledge / OpenViking Shared Skill
  → Publication Reference + Rollback Snapshot
```

### Knowledge Promotion

- WeKnora Chunk 证据重新读取并复核 KB Scope；
- 权威库语义去重与冲突检查；
- duplicate 禁止发布；冲突必须管理员显式说明；
- 审核后通过 WeKnora 正式 Manual Knowledge API 发布；
- rollback 删除本次 Promotion 创建的知识资源。

正式 Knowledge 始终只在 WeKnora。

### Skill Promotion

- Source Session / Experience 必须能由当前 OpenViking principal 真实读取；
- 候选 SKILL.md 通过 OpenViking 官方严格校验；
- 必须提供 Behavior Diff；Regression Eval 必须 `passed`；
- 修改现有 Skill 时记录 base revision，publish 前再次读取，revision 变化则阻断；
- 新建/更新只写 `viking://agent/skills`；
- rollback 支持删除本次新建 Skill 或恢复更新前内容。

正式 Skill 始终只在 OpenViking。

## 权限与安全

```text
viewer / editor → personal candidate
editor+         → workspace candidate
admin / owner   → review / publish / rollback
```

Agent Runtime 没有正式 Knowledge/Skill publish 或 rollback Tool。Promotion 使用独立 `LEECLAW_PROMOTION_TOKEN`，可信 Workspace/User/Role、WeKnora Tenant/Credential、OpenViking Account 只由 OpenClaw 服务端补齐，浏览器不能声明这些字段。

## Source of Truth

| 资产 | Source of Truth |
|---|---|
| 人类账号 / Profile | OpenClaw |
| Agent Runtime / Tool Authority | OpenClaw |
| Workspace / Role / Session Binding | LeeClaw Workspace Core |
| 正式 Knowledge / KB / RAG / Entity Graph | WeKnora |
| Ontology / Binding / Version | LeeClaw Ontology Registry |
| Memory / Session / Experience / 正式 Skill | OpenViking |
| Promotion Candidate / Review / Publication Reference | LeeClaw Promotion Store |
| 实时业务事实 | Business API / MCP / SQL Gateway |

## 代码结构

```text
cobra-knowledge/
├─ integrations/openclaw/
│  ├─ workspace-core/
│  ├─ knowledge-plugin/
│  ├─ openviking-plugin/
│  └─ promotion-plugin/
├─ internal/
│  ├─ promotion/
│  ├─ promotionapi/
│  ├─ runtimecontext/
│  ├─ retrieval/
│  ├─ context/
│  ├─ ontology/
│  └─ graphview/
├─ configs/
├─ compatibility/
└─ scripts/
```

三个 `upstream/` submodule 继续保持零修改。

## 验证

```bash
cd cobra-knowledge
./scripts/verify-v0.10.sh
./scripts/check-v0.10-upstreams.sh ../upstream/openclaw ../upstream/weknora ../upstream/openviking
```

## 当前边界

- Promotion Store 当前为单实例文件实现，v0.11 替换为共享生产 Store；
- v0.10 记录并验证 Skill Eval Result，完整自动 Eval/Trace/Replay/CI 留到 v0.12；
- Candidate 内容由上层 Agent/人工形成，Core 只做确定性治理；
- 不做无审核自动发布，不做自动覆盖正式资产。

下一版本：**v0.11 Enterprise Persistence & HA**。
