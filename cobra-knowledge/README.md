# LeeClaw Knowledge Core v0.8

> 本目录继续保留历史 `cobra-knowledge` 工程路径，避免为了命名制造大面积无关 Git diff；对外产品与插件统一使用 LeeClaw 命名。

v0.8 把已有 Knowledge Core 正式接入 OpenClaw Agent Runtime：

```text
OpenClaw Agent
├─ OpenViking Memory + Dynamic Skill Runtime
└─ LeeClaw Knowledge Tools
     ↓
   Runtime Context Service
     ↓
   Ontology / Planner / Retriever / Arbiter
     ↓
   WeKnora + Evidence
```

## 核心模块

```text
integrations/openclaw/workspace-core     Workspace / Role / Session Workspace Binding
integrations/openclaw/knowledge-plugin  Knowledge UI + Agent Knowledge Tools
integrations/openclaw/openviking-plugin Memory + Dynamic Skill Runtime
internal/runtimecontext                  在线 Context Runtime
internal/retrieval                       Semantic Catalog / Planner / Arbiter
internal/context                         Retriever / Context Pack
internal/ontology                        Ontology Registry
internal/graphview                       Entity/Ontology GraphView
internal/adapters/weknora                WeKnora 管理/检索/证据 Adapter
```

## v0.8 安全不变量

- 人类身份只来自 OpenClaw durable `profileId`；
- 已开始的 Agent Session 固定 Workspace，新 Session 才采用新的默认 Workspace；
- Skill 只能收窄 OpenClaw Tool Authority；
- Agent 不直接调用 WeKnora 管理 API；
- Evidence 返回前再次验证 Workspace KB Scope；
- Memory 不是企业事实；
- RAG fallback 不消除缺失结构化来源的 Gap；
- 三个 upstream 不改源码。

## 验证

```bash
./scripts/verify-v0.8.sh

./scripts/check-v0.8-upstreams.sh \
  ../upstream/openclaw \
  ../upstream/weknora \
  ../upstream/openviking
```

## 当前边界

v0.8 每个 Turn 自动激活一个 Skill，不执行 Skill 辅助脚本；多 KB 本体融合、通用结构化 Retriever、Promotion、HA 和完整生产 CI/E2E/Eval 留在后续版本。
