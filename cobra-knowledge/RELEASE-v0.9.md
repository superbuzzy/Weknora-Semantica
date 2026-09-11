# LeeClaw v0.9.0 Release

## 主题

**Retrieval Governance：让 Agent 查对、裁对，并明确知道是否查完整。**

v0.9.0 在 v0.8 Skill Runtime + Knowledge Runtime 基础上，直接替换原有“Planner 能选结构化来源但 Runtime 未真正配置对应 Retriever”的断层逻辑，建立完整的检索治理执行链。

## 已完成

### 1. Retrieval Planner v2

- 每个检索步骤新增 `required` 与 `freshness_requirement`；
- Plan 显式返回 `required_sources`、`supporting_sources`、`ontology_scope`、`blocking_issues`；
- 当前/实时业务指标优先选择 Business Data，并标记 `freshness=live`；
- 定义、实体事实、本体约束、诊断问题按不同来源组合生成计划；
- Planner 只负责计划，不直接访问任何业务系统。

### 2. Entity Retriever Runtime

- 新增 Runtime `EntityQuerySource`；
- WeKnora Neo4j Adapter 新增请求期实体查询；
- 支持实体定位、属性事实、关系路径；
- 实体结果保留 KB、Knowledge、Entity、Chunk 来源，可继续反查证据；
- 不修改 WeKnora upstream。

### 3. Business Retriever Contract

- 新增 transport-neutral `BusinessQuerySource`；
- 新增通用 HTTP Gateway Adapter；
- Core 不写死 SQL、数据库、具体 MCP Tool 或业务系统 URL；
- Business Data 为 required 时，未配置/失败/无结果都会保留 blocking gap。

### 4. 多 KB 本体联合

- 单请求可解析多个 KB 的 Ontology Binding；
- 每个 KB 作为独立 namespace 编译 Semantic Catalog；
- 相同 concept ID 的兼容项可合并 alias/source policy；
- 相同 label 不同 concept ID、相同 ID 不同 label 会保留 concept-sense conflict；
- 与当前 query 相关的 federation conflict 进入 blocking issue，不静默覆盖。

### 5. Evidence Contract v2

新增标准字段：

- `source_type`
- `source_id`
- `entity_id`
- `value`
- `unit`
- `effective_time`
- `confidence`
- `provenance`

同时保留 v0.8 已有字段，避免破坏 WeKnora Chunk/Graph 兼容性。

### 6. Completeness Gate

`complete=true` 不再由“有没有任意结果”决定：

- required source 未满足 → `blocking_gaps` → false；
- unresolved conflict → `blocking_gaps` → false；
- 相关本体 federation conflict → `blocking_gaps` → false；
- supporting source 失败仅记录普通 `gaps`；
- RAG fallback 只能补解释证据，不能替代缺失的实时/结构化事实。

Context Pack 新增 `source_status[]` 与 `blocking_gaps[]`，Agent 可以明确看到每个计划来源的执行结果。

## 保持不变的硬边界

- OpenClaw 仍是唯一人类账号、Agent Runtime 和最终 Tool Authority；
- WeKnora 仍是文档/RAG/Entity Graph Knowledge Engine；
- OpenViking 仍是 Memory/Experience/Skill Source of Truth；
- Ontology 生命周期仍归 LeeClaw Ontology Registry；
- 三个 upstream 源码目录保持零修改。

## 当前边界

- Business Retriever 已定义稳定接口和 HTTP Adapter，具体 MCP / SQL Gateway 由部署环境实现；
- 多 KB 本体当前完成 request-scoped Semantic Catalog federation，不做自动跨本体推理；
- Planner v2 以确定性语义规则为主，复杂计划只保留受约束 Refiner 接口；
- Workspace/Session/Audit/Ontology 持久化仍是单实例基线；
- Experience Promotion、HA、Trace/Eval/CI 按 ROADMAP 后续版本推进。

## 验证

活跃验证入口：

```bash
cd cobra-knowledge
./scripts/verify-v0.9.sh
```

上游契约：

```bash
./scripts/check-v0.9-upstreams.sh ../upstream/openclaw ../upstream/weknora ../upstream/openviking
```
