import test from "node:test";
import assert from "node:assert/strict";
import { OpenVikingClient, resolveOpenVikingConfig } from "../lib/client.js";
import { latestTurn, openVikingSessionId, renderMemoryContext } from "../lib/memory-runtime.js";
import { openVikingPrincipal, requireDurableProfileId, sessionPrincipal } from "../lib/principal.js";

function jsonResponse(body, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
}

async function withFetch(handler, fn) {
  const original = globalThis.fetch;
  globalThis.fetch = handler;
  try { await fn(); } finally { globalThis.fetch = original; }
}

const config = {
  baseUrl: "http://openviking.test",
  apiKey: "ov-key",
  workspaceId: "workspace-42",
  memoryRuntimeEnabled: true,
  memoryRecallLimit: 8,
  memoryMaxRecallChars: 12000,
  commitPendingTokens: 1200,
  commitKeepRecentCount: 6,
  requestTimeoutMs: 5000,
};
const clientIdentity = { authenticatedUserProfile: { profileId: "profile-123", displayName: "Lee" } };

test("OpenClaw durable profile is the only OpenViking user identity", () => {
  assert.equal(requireDurableProfileId(clientIdentity), "profile-123");
  assert.throws(() => requireDurableProfileId({ authenticatedUserId: "legacy-user" }), /durable OpenClaw/);
  assert.deepEqual(openVikingPrincipal(clientIdentity, config), {
    issuedBy: "openclaw",
    workspaceId: "workspace-42",
    accountId: "workspace-42",
    userId: "profile-123",
  });
});

test("v0.6 OpenViking config has no static user/account override", () => {
  const resolved = resolveOpenVikingConfig(config);
  assert.equal("userId" in resolved, false);
  assert.equal("accountId" in resolved, false);
  assert.equal("forwardAuthenticatedUser" in resolved, false);
  assert.equal(resolved.workspaceId, "workspace-42");
});

test("OpenVikingClient forwards server-owned workspace and durable profile", async () => {
  const client = new OpenVikingClient(config);
  const principal = openVikingPrincipal(clientIdentity, config);
  await withFetch(async (url, init) => {
    assert.equal(String(url), "http://openviking.test/api/v1/sessions");
    assert.equal(init.headers.Authorization, "Bearer ov-key");
    assert.equal(init.headers["X-OpenViking-Account"], "workspace-42");
    assert.equal(init.headers["X-OpenViking-User"], "profile-123");
    return jsonResponse({ result: { sessions: [] } });
  }, async () => {
    const result = await client.listSessions(principal);
    assert.deepEqual(result.sessions, []);
  });
});

test("OpenViking client fails closed without trusted user/account", () => {
  const client = new OpenVikingClient(config);
  assert.throws(() => client.headers({ accountId: "workspace-42" }), /trusted user/);
  assert.throws(() => client.headers({ userId: "profile-123" }), /trusted account/);
});

test("memory search remains memory-only and skill APIs remain OpenViking source of truth", async () => {
  const client = new OpenVikingClient(config);
  const principal = openVikingPrincipal(clientIdentity, config);
  const calls = [];
  await withFetch(async (url, init) => {
    calls.push([String(url), init.method, init.body]);
    if (String(url).endsWith("/search/find")) return jsonResponse({ result: { memories: [] } });
    if (String(url).endsWith("/skills/find")) return jsonResponse({ result: { skills: [{ name: "planning", score: 0.9 }] } });
    if (String(url).endsWith("/skills")) return jsonResponse({ result: { skills: [{ name: "planning" }] } });
    return jsonResponse({ result: { name: "planning", content: "# planning" } });
  }, async () => {
    await client.searchMemory(principal, "转供状态", 20);
    const listed = await client.listSkills(principal);
    assert.equal(listed.skills[0].name, "planning");
    const found = await client.findSkills(principal, "规划", 10);
    assert.equal(found.skills[0].score, 0.9);
    const skill = await client.getSkill(principal, "planning", "viking://agent/skills");
    assert.equal(skill.content, "# planning");
  });
  const memoryBody = JSON.parse(calls[0][2]);
  assert.equal(memoryBody.target_uri, "viking://~/memories");
  assert.equal(memoryBody.read_content, true);
});

test("runtime resolves user from OpenClaw session creation profile and never from hook params", () => {
  const api = {
    runtime: { agent: { session: { getSessionEntry: ({ sessionKey }) => sessionKey === "s-1" ? {
      createdActor: { type: "human", source: "profile", id: "profile-owner" },
    } : undefined } } },
  };
  assert.deepEqual(sessionPrincipal(api, config, "s-1"), {
    issuedBy: "openclaw-session",
    workspaceId: "workspace-42",
    accountId: "workspace-42",
    userId: "profile-owner",
  });
  assert.equal(sessionPrincipal(api, config, "missing"), null);
});

test("runtime turn parsing, session routing and memory context are deterministic", () => {
  assert.deepEqual(latestTurn([
    { role: "user", content: "问题" },
    { role: "assistant", content: [{ type: "text", text: "答案" }] },
  ]), { user: "问题", assistant: "答案" });
  const principal = openVikingPrincipal(clientIdentity, config);
  assert.equal(openVikingSessionId(principal, "session-1"), openVikingSessionId(principal, "session-1"));
  assert.notEqual(openVikingSessionId(principal, "session-1"), openVikingSessionId({ ...principal, userId: "other" }, "session-1"));
  const rendered = renderMemoryContext({ memories: [{ uri: "viking://~/memories/a", abstract: "偏好A" }] }, 5000);
  assert.match(rendered, /not authoritative enterprise facts/);
  assert.match(rendered, /偏好A/);
});

import { registerIdentityAwareMemoryRuntime } from "../lib/memory-runtime.js";

test("identity-aware memory hooks recall and capture under the session owner profile", async () => {
  const hooks = new Map();
  const calls = [];
  const api = {
    logger: { warn() {} },
    on(name, handler) { hooks.set(name, handler); },
    runtime: { agent: { session: { getSessionEntry: () => ({
      createdActor: { type: "human", source: "profile", id: "profile-owner" },
    }) } } },
  };
  const client = {
    async searchMemory(principal, query, limit) {
      calls.push(["search", principal, query, limit]);
      return { memories: [{ uri: "viking://~/memories/x", abstract: "历史偏好" }] };
    },
    async getSession(principal, id) {
      calls.push(["getSession", principal, id]);
      return { pending_tokens: 1500 };
    },
    async addSessionMessage(principal, id, role, text) {
      calls.push(["add", principal, id, role, text]);
    },
    async commitSession(principal, id, keep) {
      calls.push(["commit", principal, id, keep]);
      return { status: "accepted" };
    },
  };
  registerIdentityAwareMemoryRuntime(api, config, client);
  const promptResult = await hooks.get("before_prompt_build")({ prompt: "问题" }, { sessionKey: "s1" });
  assert.match(promptResult.prependContext, /历史偏好/);
  await hooks.get("agent_end")({ success: true, messages: [
    { role: "user", content: "问题" },
    { role: "assistant", content: "答案" },
  ] }, { sessionKey: "s1" });
  assert.equal(calls[0][1].userId, "profile-owner");
  assert.equal(calls[0][1].accountId, "workspace-42");
  assert.ok(calls.some((entry) => entry[0] === "add" && entry[3] === "user" && entry[4] === "问题"));
  assert.ok(calls.some((entry) => entry[0] === "add" && entry[3] === "assistant" && entry[4] === "答案"));
  assert.ok(calls.some((entry) => entry[0] === "commit" && entry[3] === 6));
});
