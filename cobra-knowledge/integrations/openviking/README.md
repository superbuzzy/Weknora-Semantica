# OpenViking integration in v0.6

OpenViking remains an unmodified upstream Memory + Skill Engine.

v0.6 no longer recommends the upstream OpenClaw context-engine plugin's static `accountId/userId` configuration for a multi-user LeeClaw gateway. The LeeClaw OpenViking plugin uses OpenClaw's public hook/session APIs to derive the durable session owner profile dynamically, then calls the standard OpenViking HTTP APIs with trusted Account/User headers.

No OpenViking server source changes are required.

Production requirements:

- OpenViking `server.auth_mode=trusted`;
- OpenViking reachable only from the trusted LeeClaw service network;
- `X-OpenViking-Account` generated from server-owned workspace configuration;
- `X-OpenViking-User` generated from OpenClaw durable profile id;
- root/service API key kept server-side when the OpenViking deployment requires it.
