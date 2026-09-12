# LeeClaw Knowledge Core v0.10

本目录继续保留历史 `cobra-knowledge` 工程路径，对外统一使用 LeeClaw 命名。

v0.10 在 v0.9 Retrieval Governance 上增加 `internal/promotion` 与 `internal/promotionapi`，并通过独立 OpenClaw `promotion-plugin` 提供受控候选、审核、发布和回滚。正式 Knowledge 只写 WeKnora，正式 Skill 只写 OpenViking；Agent Runtime 不拥有发布类 Tool。

```text
OpenClaw Agent Runtime
├─ Memory / Experience / Skill Runtime
├─ Retrieval Governance
└─ Promotion Management Plane
   ├─ Candidate / Inspection / Review
   ├─ Promotion Gate
   ├─ WeKnora Knowledge Publisher
   └─ OpenViking Shared Skill Publisher / Rollback
```

验证：`./scripts/verify-v0.10.sh`。

当前 Promotion Store 为单实例 `FSStore`；共享持久化与 HA 属于 v0.11，完整 Trace/Eval/CI 属于 v0.12。
