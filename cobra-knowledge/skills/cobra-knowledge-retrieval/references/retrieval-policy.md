# CobraKnowledge v0.2 检索策略契约

## Source

- `business_data`：实时/事务数据库或 API。
- `entity_graph`：具体实体、属性、状态、关系与路径。
- `wiki_rag`：文档知识、制度、原因、背景和原文证据。
- `ontology_graph`：类型、属性、关系、层级和约束。

## RetrievalPlan

必须包含 Query、Semantic、Steps、并行组、是否需要 Arbitration、是否需要 Evidence、Stop Condition。复杂问题可标记 `requires_llm_planning=true`，由受约束的 Complex Planner 补充，不改变硬策略。

## Assertion

事实与实体分离。Assertion 至少包含 subject、predicate、value、source、authority、confidence；可进一步带 scope、observed_at、valid_from、valid_to、version、supersedes、evidence。

## 硬判断

来源优先级、有效时间、过期 TTL、版本状态、Domain/Range、权限和写入门禁由代码执行。

## 软判断

意图拆解、同义表达、复杂多跳计划、不同文本是否描述同一事实槽位可交给 LLM，但其结果必须进入确定性策略层执行。
