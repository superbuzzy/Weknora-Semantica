function required(value, name) {
  const text = String(value ?? "").trim();
  if (!text) throw new Error(`leeclaw-context: ${name} is required`);
  return text;
}

export class ContextRuntimeClient {
  constructor(config) { this.config = config; }

  headers(principal) {
    const headers = {
      Accept: "application/json",
      "Content-Type": "application/json",
      Authorization: `Bearer ${required(this.config.runtimeToken, "runtime token")}`,
      "X-LeeClaw-Workspace-ID": required(principal.workspaceId, "workspace id"),
      "X-LeeClaw-User-ID": required(principal.userId, "user id"),
      "X-LeeClaw-WeKnora-Tenant-ID": required(principal.tenantId, "WeKnora tenant"),
      "X-LeeClaw-WeKnora-API-Key": required(principal.weknoraApiKey, "WeKnora workspace credential"),
    };
    if (principal.weknoraBaseUrl) headers["X-LeeClaw-WeKnora-Base-URL"] = String(principal.weknoraBaseUrl);
    return headers;
  }

  async request(path, principal, body) {
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), this.config.requestTimeoutMs);
    try {
      const response = await fetch(`${required(this.config.coreApiBaseUrl, "core API base URL")}${path}`, {
        method: "POST",
        headers: this.headers(principal),
        body: JSON.stringify(body),
        signal: controller.signal,
      });
      const text = await response.text();
      let payload = text;
      if (text) { try { payload = JSON.parse(text); } catch { /* keep text */ } }
      if (!response.ok) {
        const detail = typeof payload === "string" ? payload : JSON.stringify(payload);
        throw new Error(`LeeClaw Core POST ${path} -> ${response.status}: ${detail.slice(0, 1200)}`);
      }
      return payload;
    } finally { clearTimeout(timer); }
  }

  retrieve(principal, input, knowledgeBaseIds) {
    return this.request("/api/v1/runtime/context/retrieve", principal, {
      query: required(input?.query, "query"),
      domain: String(input?.domain ?? "").trim(),
      task: String(input?.task ?? "").trim(),
      knowledge_base_ids: knowledgeBaseIds,
    });
  }

  getEvidence(principal, chunkId) {
    return this.request("/api/v1/runtime/context/evidence", principal, {
      chunk_id: required(chunkId, "chunk id"),
      knowledge_base_ids: Array.isArray(principal?.knowledgeBaseIds) ? principal.knowledgeBaseIds : [],
    });
  }
}
