import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { AuditLog } from "../lib/audit-log.js";

test("audit log is workspace-scoped and can filter a knowledge base including child resources", () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "leeclaw-audit-"));
  const file = path.join(dir, "audit.jsonl");
  const audit = new AuditLog(file);
  audit.record({ actorProfileId: "p1", workspaceId: "cq", action: "knowledge_base.update", resourceType: "knowledge_base", resourceId: "kb1" });
  audit.record({ actorProfileId: "p1", workspaceId: "cq", action: "document.reparse", resourceType: "document", resourceId: "doc1", detail: { knowledgeBaseId: "kb1" } });
  audit.record({ actorProfileId: "p2", workspaceId: "hq", action: "knowledge_base.update", resourceType: "knowledge_base", resourceId: "kb1" });
  const rows = audit.list("cq", { resourceId: "kb1" });
  assert.equal(rows.length, 2);
  assert.deepEqual(rows.map((x) => x.action), ["document.reparse", "knowledge_base.update"]);
  assert.equal(fs.statSync(file).mode & 0o777, 0o600);
});

test("audit detail is allow-listed and never persists arbitrary payloads", () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), "leeclaw-audit-"));
  const audit = new AuditLog(path.join(dir, "audit.jsonl"));
  audit.record({ actorProfileId: "p1", workspaceId: "cq", action: "document.upload", detail: { knowledgeBaseId: "kb1", base64: "SECRET", content: "SECRET", name: "a.txt" } });
  const [row] = audit.list("cq");
  assert.equal(row.detail.knowledgeBaseId, "kb1");
  assert.equal(row.detail.name, "a.txt");
  assert.equal("base64" in row.detail, false);
  assert.equal("content" in row.detail, false);
});
