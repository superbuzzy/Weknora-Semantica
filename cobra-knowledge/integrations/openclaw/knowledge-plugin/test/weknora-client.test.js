import test from "node:test";
import assert from "node:assert/strict";
import { resolveKnowledgeConfig } from "../lib/config.js";
import { knowledgePrincipal, requireDurableProfileId } from "../lib/principal.js";
import { WeKnoraClient } from "../lib/weknora-client.js";

function jsonResponse(body, status = 200) {
  return new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
}

async function withFetch(handler, fn) {
  const original = globalThis.fetch;
  globalThis.fetch = handler;
  try { await fn(); } finally { globalThis.fetch = original; }
}

const config = {
  weknoraBaseUrl: "http://weknora.test",
  graphApiBaseUrl: "http://graph.test",
  weknoraApiKey: "service-key",
  weknoraTenantId: "42",
  workspaceId: "workspace-42",
  requestTimeoutMs: 5000,
};

const clientIdentity = {
  authenticatedUserId: "email@example.com",
  authenticatedUserProfile: { profileId: "profile-123", displayName: "Lee" },
};

test("durable OpenClaw profile id is the only accepted user identity", () => {
  assert.equal(requireDurableProfileId(clientIdentity), "profile-123");
  assert.throws(() => requireDurableProfileId({ authenticatedUserId: "email@example.com" }), /durable OpenClaw/);
  assert.deepEqual(knowledgePrincipal(clientIdentity, config), {
    issuedBy: "openclaw",
    userId: "profile-123",
    externalUserId: "profile-123",
    workspaceId: "workspace-42",
    tenantId: "42",
  });
});

test("v0.6 config rejects legacy user and bearer-token identity configuration", () => {
  assert.throws(() => resolveKnowledgeConfig({ weknoraBaseUrl: "http://x", weknoraApiKey: "k", weknoraTenantId: "1" }), /workspaceId/);
  const resolved = resolveKnowledgeConfig(config);
  assert.equal("weknoraBearerToken" in resolved, false);
  assert.equal("externalUserId" in resolved, false);
  assert.equal("forwardAuthenticatedUser" in resolved, false);
  assert.equal("tenantId" in resolved, false);
});

test("WeKnoraClient uses service API key plus server-issued tenant and OpenClaw principal", async () => {
  const client = new WeKnoraClient(config);
  const principal = knowledgePrincipal(clientIdentity, config);
  await withFetch(async (url, init) => {
    assert.equal(String(url), "http://weknora.test/api/v1/knowledge-bases");
    assert.equal(init.headers.Authorization, undefined);
    assert.equal(init.headers["X-API-Key"], "service-key");
    assert.equal(init.headers["X-Tenant-ID"], "42");
    assert.equal(init.headers["X-External-User-ID"], "profile-123");
    return jsonResponse({ data: [{ id: "kb-1" }] });
  }, async () => {
    const result = await client.listKnowledgeBases(principal);
    assert.equal(result[0].id, "kb-1");
  });
});

test("WeKnoraClient fails closed when trusted downstream identity is absent", () => {
  const client = new WeKnoraClient(config);
  assert.throws(() => client.headers({ tenantId: "42" }), /OpenClaw principal/);
  assert.throws(() => client.headers({ externalUserId: "profile-123" }), /WeKnora tenant/);
});

test("ontology graph keeps the same GraphView endpoint and principal", async () => {
  const client = new WeKnoraClient(config);
  const principal = knowledgePrincipal(clientIdentity, config);
  await withFetch(async (url, init) => {
    assert.equal(String(url), "http://graph.test/api/v1/knowledge-bases/kb-1/graph?view=ontology&limit=180");
    assert.equal(init.headers["X-External-User-ID"], "profile-123");
    return jsonResponse({ nodes: [{ id: "class:line" }], edges: [], meta: { view: "ontology" } });
  }, async () => {
    const result = await client.graph(principal, "kb-1", "ontology");
    assert.equal(result.meta.view, "ontology");
  });
});
