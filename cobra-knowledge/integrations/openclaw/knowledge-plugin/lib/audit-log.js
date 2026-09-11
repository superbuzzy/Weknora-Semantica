import fs from "node:fs";
import path from "node:path";

function required(value, name) {
  const text = String(value ?? "").trim();
  if (!text) throw new Error(`leeclaw-audit: ${name} is required`);
  return text;
}

function safeDetail(value) {
  if (!value || typeof value !== "object") return undefined;
  const allowed = {};
  for (const key of ["role", "permission", "name", "targetProfileId", "organizationId", "targetWorkspaceId", "knowledgeBaseId", "status"]) {
    const item = value[key];
    if (item !== undefined && item !== null && String(item).length <= 256) allowed[key] = item;
  }
  return Object.keys(allowed).length ? allowed : undefined;
}

export class AuditLog {
  constructor(filePath) {
    this.filePath = required(filePath, "auditLogPath");
  }

  record(event) {
    const entry = {
      at: new Date().toISOString(),
      actorProfileId: required(event?.actorProfileId, "actorProfileId"),
      workspaceId: required(event?.workspaceId, "workspaceId"),
      action: required(event?.action, "action"),
      outcome: event?.outcome === "failure" ? "failure" : "success",
      ...(event?.resourceType ? { resourceType: String(event.resourceType) } : {}),
      ...(event?.resourceId ? { resourceId: String(event.resourceId) } : {}),
      ...(event?.error ? { error: String(event.error).slice(0, 500) } : {}),
      ...(safeDetail(event?.detail) ? { detail: safeDetail(event.detail) } : {}),
    };
    fs.mkdirSync(path.dirname(this.filePath), { recursive: true });
    fs.appendFileSync(this.filePath, `${JSON.stringify(entry)}\n`, { encoding: "utf8", mode: 0o600 });
    try { fs.chmodSync(this.filePath, 0o600); } catch { /* best effort */ }
    return entry;
  }

  list(workspaceId, { limit = 100, resourceId } = {}) {
    const target = required(workspaceId, "workspaceId");
    const size = Number.isFinite(Number(limit)) ? Math.max(1, Math.min(Number(limit), 500)) : 100;
    const resource = String(resourceId ?? "").trim();
    let text = "";
    try { text = fs.readFileSync(this.filePath, "utf8"); }
    catch (error) { if (error?.code === "ENOENT") return []; throw error; }
    const rows = text.split("\n").filter(Boolean).flatMap((line) => {
      try {
        const item = JSON.parse(line);
        if (item?.workspaceId !== target) return [];
        if (resource && String(item?.resourceId ?? "") !== resource && String(item?.detail?.knowledgeBaseId ?? "") !== resource) return [];
        return [item];
      }
      catch { return []; }
    });
    return rows.slice(-size).reverse();
  }
}
