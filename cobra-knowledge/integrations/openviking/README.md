# OpenViking integration in v0.7

OpenViking is LeeClaw's **Memory + Skill Engine**.

v0.7 does not use a static OpenViking account/user for a multi-user OpenClaw gateway. The LeeClaw OpenViking plugin resolves the durable OpenClaw session/profile identity and the validated LeeClaw Workspace at runtime, then calls standard OpenViking HTTP APIs with trusted Account/User headers.

```text
OpenClaw profileId
  + LeeClaw Workspace selection
  → OpenViking User / Account
```

The Workspace resolver does not read WeKnora credentials, so OpenViking remains independent from the Knowledge Engine.

Memory lifecycle uses OpenClaw public hooks; Skill remains stored only in OpenViking. No OpenViking upstream source file is modified.
