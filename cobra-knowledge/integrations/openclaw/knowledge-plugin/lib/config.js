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

  const requestTimeoutMs = Number(raw.requestTimeoutMs ?? process.env.LEECLAW_KNOWLEDGE_TIMEOUT_MS ?? 15000);
  return {
    weknoraBaseUrl,
    graphApiBaseUrl: cleanUrl(raw.graphApiBaseUrl ?? process.env.LEECLAW_GRAPH_API_BASE_URL),
    weknoraApiKey: requiredString(raw.weknoraApiKey ?? process.env.WEKNORA_API_KEY, "weknoraApiKey"),
    weknoraTenantId: requiredString(raw.weknoraTenantId ?? process.env.WEKNORA_TENANT_ID, "weknoraTenantId"),
    workspaceId: requiredString(raw.workspaceId ?? process.env.LEECLAW_WORKSPACE_ID, "workspaceId"),
    requestTimeoutMs: Number.isFinite(requestTimeoutMs)
      ? Math.max(1000, Math.min(requestTimeoutMs, 60000))
      : 15000,
  };
}
