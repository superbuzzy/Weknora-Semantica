# CobraKnowledge v0.2 本体契约

## 三图

- Wiki 图：知识主题/页面关系，由 WeKnora 维护。
- 实体图：具体实体、属性、关系和事实，由候选图逐步治理为正式业务事实图。
- 本体图：Class、Property、Relation、Hierarchy、Domain/Range、Constraint，不承载具体业务事实值。

## 核心对象

`OntologyClass`、`DataProperty`、`ObjectRelation`、`Constraint`、`ReviewItem`、`Entity`、`Relation`、`Assertion`、`EvidenceRef`。

## 生命周期

候选本体：`candidate`；审核发布：`approved`；废弃：`superseded`；拒绝：`rejected`。已批准版本不可原地覆盖，变更创建新版本。

## 机器 ID 与中文业务名

机器 ID 使用稳定哈希生成，不依赖英文翻译或中文转拼音。`label` 始终保存权威业务中文名称，避免第三方命名规范影响业务语义。

## 高风险审核项

类型合并/拆分、顶层类变更、等价/互斥、删除、强制基数、跨域复用、关系方向大范围调整必须人工审核。
