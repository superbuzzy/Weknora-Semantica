import test from "node:test";
import assert from "node:assert/strict";
import { OpenVikingClient } from "../lib/client.js";

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

test("OpenVikingClient forwards trusted account/user and API key", async () => {
  const client = new OpenVikingClient({
    baseUrl: "http://openviking.test",
    apiKey: "ov-key",
    accountId: "workspace-42",
    requestTimeoutMs: 5000,
  });

  await withFetch(async (url, init) => {
    assert.equal(String(url), "http://openviking.test/api/v1/sessions");
    assert.equal(init.headers.Authorization, "Bearer ov-key");
    assert.equal(init.headers["X-OpenViking-Account"], "workspace-42");
    assert.equal(init.headers["X-OpenViking-User"], "user-123");
    return jsonResponse({ status: "ok", result: { sessions: [] } });
  }, async () => {
    const result = await client.listSessions({ userId: "user-123" });
    assert.deepEqual(result.sessions, []);
  });
});

test("OpenVikingClient searches user memory without mixing enterprise resources", async () => {
  const client = new OpenVikingClient({ baseUrl: "http://openviking.test", requestTimeoutMs: 5000 });

  await withFetch(async (url, init) => {
    assert.equal(String(url), "http://openviking.test/api/v1/search/find");
    const body = JSON.parse(init.body);
    assert.equal(body.query, "转供状态");
    assert.equal(body.target_uri, "viking://~/memories");
    assert.equal(body.read_content, true);
    return jsonResponse({ status: "ok", result: { memories: [] } });
  }, async () => {
    await client.searchMemory({}, "转供状态", 20);
  });
});

test("OpenVikingClient lists, finds and reads SKILL.md through the Skills API", async () => {
  const client = new OpenVikingClient({ baseUrl: "http://openviking.test", requestTimeoutMs: 5000 });
  const calls = [];

  await withFetch(async (url, init) => {
    calls.push([String(url), init.method, init.body]);
    if (String(url).endsWith("/api/v1/skills")) {
      return jsonResponse({ status: "ok", result: { skills: [{ name: "planning" }] } });
    }
    if (String(url).endsWith("/api/v1/skills/find")) {
      return jsonResponse({ status: "ok", result: { skills: [{ name: "planning", score: 0.9 }] } });
    }
    return jsonResponse({ status: "ok", result: { name: "planning", content: "# planning" } });
  }, async () => {
    const listed = await client.listSkills({});
    assert.equal(listed.skills[0].name, "planning");
    const found = await client.findSkills({}, "规划", 10);
    assert.equal(found.skills[0].score, 0.9);
    const skill = await client.getSkill({}, "planning", "viking://agent/skills");
    assert.equal(skill.content, "# planning");
  });

  assert.equal(calls[1][0], "http://openviking.test/api/v1/skills/find");
  assert.deepEqual(JSON.parse(calls[1][2]), { query: "规划", limit: 10, level: [0, 1] });
  assert.equal(calls[2][0], "http://openviking.test/api/v1/skills/planning?level=2&include_content=true&include_files=true&target_uri=viking%3A%2F%2Fagent%2Fskills");
});
