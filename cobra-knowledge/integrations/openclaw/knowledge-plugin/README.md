# leeclaw-knowledge v0.7

OpenClaw 原生 Knowledge + Workspaces Feature Plugin。

## 作用

- 用 OpenClaw durable `profileId` 作为唯一人类身份；
- 使用服务端 Workspace Registry 做成员、角色、当前 Workspace 和下游映射；
- 将 WeKnora KB / Document / Wiki / FAQ / Tag / Search / Share 能力原生呈现在 OpenClaw；
- 展示 Entity Graph / Ontology Graph；
- 记录 LeeClaw Knowledge/Workspace 操作审计。

## 权限

```text
OpenClaw operator scope
AND Workspace role
AND WeKnora service capability/tenant boundary
```

浏览器只能请求逻辑 Workspace ID，不能提交下游 Tenant/Account/User/Credential。

## 配置

见：

- `configs/openclaw-v0.7.example.json`
- `configs/workspaces-v0.7.example.json`

Workspace 文件中的 `weknora.apiKeyEnv` 仅是 Secret 环境变量名；Key 本身不写入 Registry。

## 上传边界

当前 Control UI 文件上传使用 base64 RPC，默认 10 MiB、硬上限 50 MiB。后续大文件应改 authenticated streaming route。
