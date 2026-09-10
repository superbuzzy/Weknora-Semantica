---
name: cobra-knowledge-retrieval
description: 用于企业 Agent 的多源可信检索：依据本体语义在 Wiki RAG、实体图、本体图和业务数据之间规划最小充分检索路径，处理知识时效、范围、来源权威、版本与冲突，并生成可追溯 Context Pack。适用于事实查询、问数、诊断、规划分析、规则解释和跨源证据检索。
---

# Cobra Knowledge Retrieval

Skill 是业务方法和策略源；Planner、Arbiter、Retriever 是确定性执行框架。不要让模型每轮自行遍历所有数据源。

## 工作流

1. 明确任务：事实、实时数据、解释/诊断、Schema、本体或混合任务。
2. 提取显式范围：实体、时间、组织、区域、业务域、属性、关系。
3. 优先调用统一 `context.retrieve`，由 Context Service 完成规划和并行检索。
4. 默认检索职责：
   - 具体实体/属性/关系/路径：实体图；
   - 实时或事务指标：属性策略指定的业务数据源；
   - 原因、制度、依据、报告解释：Wiki/RAG；
   - Class/Property/Relation、Domain/Range、合法路径：本体图。
5. 独立来源并行执行，不机械串行查三张图。
6. 关键事实进入 Arbiter；LLM 不直接决定哪个冲突值胜出。
7. 结论必须满足证据门槛；证据不足时扩展检索，不补造事实。
8. 达到 Stop Condition 后立即停止，避免无意义继续搜索。

## 冲突与过期

- 先判断是否同一事实槽位：主体、谓词、时间、范围、口径必须一致。
- `valid_from/valid_to` 决定事实是否适用于查询时间；“新”不等于“对历史问题更准确”。
- 新鲜度、来源优先级、版本和权威性按 Policy 由代码执行。
- Arbiter 返回 `unresolved_conflict` 时必须保留冲突，不强行生成唯一事实。
- 未解决冲突影响写操作时，阻断自动执行并进入审批/人工确认。

## 边界

Skill 不保存具体业务事实，不写 SQL/Cypher/API 路径。底层数据源变化只改 Adapter/Retriever。

字段说明见 `references/retrieval-policy.md`。
