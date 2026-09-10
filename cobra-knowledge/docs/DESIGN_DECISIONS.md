# v0.2 关键设计决策

1. **两上游零侵入**：WeKnora 和 Semantica 均不修改；CobraKnowledge 不依赖 Semantica 包。
2. **Go 原生核心**：在线高频 Planner/Arbiter/Context Service 与离线 Ontology Discovery 使用同一 Go 数据契约。
3. **不追求先做完整 OWL/SHACL 平台**：先满足 Agent 检索、约束、证据、冲突治理；未来可增加 RDF/SHACL 导出适配器。
4. **中文业务名与机器 ID 分离**：避免英文命名/转写成为语义资产。
5. **事实不覆盖**：Entity 解决身份，Assertion 解决可变化事实，Evidence 解决证据。
6. **LLM 软、代码硬**：语义归纳/复杂计划可用 LLM；来源优先级、时间、Domain/Range、权限等由代码执行。
7. **最小充分检索**：Planner 不默认查三张图；并发仅发生在确有多源需求时。
8. **Agent 工具收敛**：生产尽量只暴露 `context.retrieve` / `context.get_evidence`，底层工具用于治理与调试。
