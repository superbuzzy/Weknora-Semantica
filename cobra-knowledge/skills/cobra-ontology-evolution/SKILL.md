---
name: cobra-ontology-evolution
description: 用于从 WeKnora 候选实体图发现、审核、发布和演进企业业务本体，并用正式本体反向约束后续实体图构建。适用于本体发现、类型属性关系归纳、Domain Range 审核、实体归一、候选概念治理、本体版本升级及三图知识体系建设。CobraKnowledge 自研实现，不依赖 Semantica 运行时。
---

# Cobra Ontology Evolution

采用“Bottom-up 发现 + Top-down 治理”闭环。WeKnora 与 Semantica 均视为上游/方法参考，不修改其源码，不把 Semantica 作为运行依赖。

## 首次领域建模

1. 从 WeKnora 候选实体图读取实体、属性、关系及 Chunk 证据。
2. 调用 CobraKnowledge Graph Normalizer，保留业务中文名称，生成稳定机器 ID。
3. 运行 Pattern Analyzer，统计类型、属性、关系、Domain/Range 组合及支持度。
4. 只有在需要同义词、上下位关系、类型合并建议等语义判断时，使用 `prompts/ontology-semantic-induction.md` 的约束进行 LLM 归纳。
5. 将归纳结果生成 Candidate Ontology；Class merge、顶层类调整、跨域复用等高风险项必须进入 Review Queue。
6. 运行 Ontology Validator。存在 error 时不得发布。
7. 业务专家审核后发布不可变 Ontology Version。

## 正式本体后的持续构图

1. 将已批准 Ontology 编译为 WeKnora `ExtractConfig` 与中文抽取约束。
2. 新实体、属性、关系必须满足已批准 Schema；未知概念进入候选区，不自动扩充正式本体。
3. 构图后先做保守实体归一，再做 Domain/Range、属性归属等确定性校验。
4. 实体属性值和关系事实保存为 Assertion，并绑定 Evidence；不要把“当前值”直接覆盖进唯一实体事实。
5. 文档更新或删除时更新 Evidence/Assertion 生命周期，不盲删共享实体。

## 边界

- Skill 保存业务建模方法和审核原则，不写 Neo4j Cypher、WeKnora HTTP 路径或存储细节。
- 不因名称相似自动合并类型或实体。
- 不让 LLM 自动批准高风险本体变更。
- 不把本体图、实体图、Wiki 图合并成同一种数据模型。

详细模型与审核门禁见 `references/ontology-contract.md`。
