import path from "node:path";

function cleanUrl(value) {
  return String(value ?? "").trim().replace(/\/+$/u, "");
}

function requiredString(value, name) {
  const text = String(value ?? "").trim();
  if (!text) throw new Error(`leeclaw-knowledge: ${name} is required`);
  return text;
}

export function resolveKnowledgeConfig(raw = {}) {
  const weknoraBaseUrl = cleanUrl(raw.weknoraBaseUrl ?? process.env.WEKNORA_BASE_URL);
  if (!weknoraBaseUrl) throw new Error("leeclaw-knowledge: weknoraBaseUrl is required");
  const workspaceRegistryPath = requiredString(raw.workspaceRegistryPath ?? process.env.LEECLAW_WORKSPACE_REGISTRY, "workspaceRegistryPath");
  const workspaceStatePath = String(raw.workspaceStatePath ?? process.env.LEECLAW_WORKSPACE_STATE ?? "").trim() || undefined;
  const defaultStatePath = workspaceStatePath ?? `${workspaceRegistryPath}.state.json`;
  const requestTimeoutMs = Number(raw.requestTimeoutMs ?? process.env.LEECLAW_KNOWLEDGE_TIMEOUT_MS ?? 30000);
  const uploadMaxBytes = Number(raw.uploadMaxBytes ?? process.env.LEECLAW_UPLOAD_MAX_BYTES ?? 10 * 1024 * 1024);
  return {
    weknoraBaseUrl,
    graphApiBaseUrl: cleanUrl(raw.graphApiBaseUrl ?? process.env.LEECLAW_GRAPH_API_BASE_URL),
    workspaceRegistryPath,
    workspaceStatePath,
    auditLogPath: requiredString(raw.auditLogPath ?? process.env.LEECLAW_AUDIT_LOG ?? path.join(path.dirname(defaultStatePath), "knowledge-audit.jsonl"), "auditLogPath"),
    requestTimeoutMs: Number.isFinite(requestTimeoutMs) ? Math.max(1000, Math.min(requestTimeoutMs, 120000)) : 30000,
    uploadMaxBytes: Number.isFinite(uploadMaxBytes) ? Math.max(1024, Math.min(uploadMaxBytes, 50 * 1024 * 1024)) : 10 * 1024 * 1024,
  };
}
