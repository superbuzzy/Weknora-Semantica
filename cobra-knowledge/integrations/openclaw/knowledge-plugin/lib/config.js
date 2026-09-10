function cleanUrl(value) {
  return String(value ?? "").trim().replace(/\/+$/u, "");
}

function cleanString(value) {
  const text = String(value ?? "").trim();
  return text || undefined;
}

export function resolveKnowledgeConfig(raw = {}) {
  const weknoraBaseUrl = cleanUrl(raw.weknoraBaseUrl ?? process.env.WEKNORA_BASE_URL);
  if (!weknoraBaseUrl) {
    throw new Error("leeclaw-knowledge: weknoraBaseUrl is required");
  }
  const requestTimeoutMs = Number(raw.requestTimeoutMs ?? process.env.LEECLAW_KNOWLEDGE_TIMEOUT_MS ?? 15000);
  return {
    weknoraBaseUrl,
    graphApiBaseUrl: cleanUrl(raw.graphApiBaseUrl ?? process.env.LEECLAW_GRAPH_API_BASE_URL),
    weknoraBearerToken: cleanString(raw.weknoraBearerToken ?? process.env.WEKNORA_BEARER_TOKEN),
    weknoraApiKey: cleanString(raw.weknoraApiKey ?? process.env.WEKNORA_API_KEY),
    tenantId: cleanString(raw.tenantId ?? process.env.WEKNORA_TENANT_ID),
    externalUserId: cleanString(raw.externalUserId ?? process.env.WEKNORA_EXTERNAL_USER_ID),
    forwardAuthenticatedUser: raw.forwardAuthenticatedUser !== false,
    requestTimeoutMs: Number.isFinite(requestTimeoutMs) ? Math.max(1000, Math.min(requestTimeoutMs, 60000)) : 15000,
  };
}
