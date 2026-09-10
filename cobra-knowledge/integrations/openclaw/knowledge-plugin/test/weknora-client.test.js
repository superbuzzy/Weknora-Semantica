import test from "node:test";
import assert from "node:assert/strict";
import { WeKnoraClient } from "../lib/weknora-client.js";

function jsonResponse(body, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

async function withFetch(handler, fn) {
  const original = globalThis.fetch;
  globalThis.fetch = handler;
  try { await fn(); } finally { globalThis.fetch = original; }
}

test("WeKnoraClient forwards bearer, tenant and external identity", async () => {
  const client = new WeKnoraClient({
    weknoraBaseUrl: "http://weknora.test",
    graphApiBaseUrl: "http://graph.test",
    weknoraBearerToken: "token-1",
    tenantId: "42",
    requestTimeoutMs: 5000,
  });

  await withFetch(async (url, init) => {
    assert.equal(String(url), "http://weknora.test/api/v1/knowledge-bases");
    assert.equal(init.headers.Authorization, "Bearer token-1");
    assert.equal(init.headers["X-Tenant-ID"], "42");
    assert.equal(init.headers["X-External-User-ID"], "user-123");
    return jsonResponse({ data: [{ id: "kb-1" }] });
  }, async () => {
    const result = await client.listKnowledgeBases({ externalUserId: "user-123" });
    assert.equal(result[0].id, "kb-1");
  });
});

test("WeKnoraClient uses API key when bearer is absent", async () => {
  const client = new WeKnoraClient({
    weknoraBaseUrl: "http://weknora.test",
    weknoraApiKey: "api-key-1",
    requestTimeoutMs: 5000,
  });

  await withFetch(async (_url, init) => {
    assert.equal(init.headers["X-API-Key"], "api-key-1");
    assert.equal(init.headers.Authorization, undefined);
    return jsonResponse({ data: [] });
  }, async () => {
    await client.listKnowledgeBases({});
  });
});

test("WeKnoraClient keeps ontology graph behind the stable GraphView endpoint", async () => {
  const client = new WeKnoraClient({
    weknoraBaseUrl: "http://weknora.test",
    graphApiBaseUrl: "http://graph.test",
    weknoraApiKey: "api-key-1",
    tenantId: "42",
    requestTimeoutMs: 5000,
  });

  await withFetch(async (url, init) => {
    assert.equal(String(url), "http://graph.test/api/v1/knowledge-bases/kb-1/graph?view=ontology&limit=180");
    assert.equal(init.headers["X-API-Key"], "api-key-1");
    assert.equal(init.headers["X-Tenant-ID"], "42");
    return jsonResponse({ nodes: [{ id: "class:line" }], edges: [], meta: { view: "ontology" } });
  }, async () => {
    const result = await client.graph({}, "kb-1", "ontology");
    assert.equal(result.meta.view, "ontology");
  });
});
