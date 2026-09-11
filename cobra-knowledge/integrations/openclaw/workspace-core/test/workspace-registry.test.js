import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { WorkspaceRegistry, sessionBindingKey } from "../index.js";

function fixture() {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "leeclaw-ws-"));
  const registryPath = path.join(dir, "workspaces.json");
  const statePath = path.join(dir, "state.json");
  fs.writeFileSync(registryPath, JSON.stringify({ version: 1, workspaces: [
    { id: "cq", name: "重庆公司", members: [{ profileId: "p1", role: "owner" }, { profileId: "p2", role: "viewer" }], weknora: { tenantId: "42", apiKeyEnv: "WK_CQ" }, openviking: { accountId: "cq" } },
    { id: "hq", name: "总部", members: [{ profileId: "p1", role: "editor" }], weknora: { tenantId: "1", apiKeyEnv: "WK_HQ" }, openviking: { accountId: "hq" } }
  ] }), "utf8");
  process.env.WK_CQ = "key-cq"; process.env.WK_HQ = "key-hq";
  return { registryPath, statePath, registry: new WorkspaceRegistry(registryPath, statePath) };
}

function apiFor(profileId) {
  return { runtime: { agent: { session: { getSessionEntry() { return { createdActor: { type: "human", source: "profile", id: profileId } }; } } } } };
}

test("workspace resolver only exposes memberships and persists validated default switching in canonical v0.8 state", () => {
  const { registry, statePath } = fixture();
  assert.deepEqual(registry.listFor("p2").map(x => x.id), ["cq"]);
  assert.equal(registry.resolve("p1").id, "cq");
  assert.equal(registry.switch("p1", "hq").weknoraTenantId, "1");
  assert.equal(registry.resolve("p1").id, "hq");
  const state = JSON.parse(fs.readFileSync(statePath, "utf8"));
  assert.equal(state.version, 2);
  assert.equal(state.profiles.p1, "hq");
  assert.deepEqual(state.sessions, {});
  assert.throws(() => registry.switch("p2", "hq"), /not accessible/);
});

test("workspace roles are server-enforced", () => {
  const { registry } = fixture();
  assert.equal(registry.requireRole("p1", "cq", "admin").role, "owner");
  assert.throws(() => registry.requireRole("p2", "cq", "editor"), /editor role is required/);
});

test("workspace owner can map OpenClaw profiles without creating downstream accounts", () => {
  const { registry } = fixture();
  registry.upsertMember("p1", "cq", "p3", "editor", "User 3");
  assert.equal(registry.resolve("p3", "cq").role, "editor");
  registry.removeMember("p1", "cq", "p3");
  assert.throws(() => registry.resolve("p3", "cq"), /no workspace membership/);
});

test("workspace registry never leaves a workspace without an owner", () => {
  const { registry } = fixture();
  assert.throws(() => registry.upsertMember("p1", "cq", "p1", "admin", "Owner"), /last owner/);
  assert.throws(() => registry.removeMember("p1", "cq", "p1"), /remove self|last owner/);
});

test("agent session is pinned to the workspace selected when that session first runs", () => {
  const { registry, statePath } = fixture();
  const api = apiFor("p1");
  assert.equal(registry.resolveSession(api, "session-key", "session-id-1").id, "cq");
  registry.switch("p1", "hq");
  assert.equal(registry.resolve("p1").id, "hq");
  assert.equal(registry.resolveSession(api, "session-key", "session-id-1").id, "cq");
  assert.equal(registry.resolveSession(api, "session-key", "session-id-2").id, "hq");

  const state = JSON.parse(fs.readFileSync(statePath, "utf8"));
  assert.equal(state.sessions[sessionBindingKey("session-id-1", "session-key")].workspaceId, "cq");
  assert.equal(state.sessions[sessionBindingKey("session-id-2", "session-key")].workspaceId, "hq");
  assert.equal(JSON.stringify(state).includes("session-id-1"), false);
});

test("session binding fails closed after its profile loses workspace membership", () => {
  const { registry } = fixture();
  const p2api = apiFor("p2");
  assert.equal(registry.resolveSession(p2api, "s-p2", "sid-p2").id, "cq");
  registry.removeMember("p1", "cq", "p2");
  assert.equal(registry.resolveSession(p2api, "s-p2", "sid-p2"), null);
});

test("workspace exposes optional Knowledge scope through the same shared core", () => {
  const { registry } = fixture();
  assert.deepEqual(registry.resolve("p1", "cq").weknoraKnowledgeBaseIds, []);
});
