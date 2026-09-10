function cleanUrl(value) { return String(value ?? "").trim().replace(/\/+$/u, ""); }
function clean(value) { const v = String(value ?? "").trim(); return v || undefined; }

export function resolveOpenVikingConfig(raw = {}) {
  const baseUrl = cleanUrl(raw.baseUrl ?? process.env.OPENVIKING_BASE_URL);
  if (!baseUrl) throw new Error("leeclaw-openviking: baseUrl is required");
  const timeout = Number(raw.requestTimeoutMs ?? 15000);
  return {
    baseUrl,
    apiKey: clean(raw.apiKey ?? process.env.OPENVIKING_API_KEY),
    accountId: clean(raw.accountId ?? process.env.OPENVIKING_ACCOUNT_ID),
    userId: clean(raw.userId ?? process.env.OPENVIKING_USER_ID),
    forwardAuthenticatedUser: raw.forwardAuthenticatedUser !== false,
    requestTimeoutMs: Number.isFinite(timeout) ? Math.max(1000, Math.min(timeout, 60000)) : 15000,
  };
}

export class OpenVikingClient {
  constructor(config) { this.config = config; }

  headers(context = {}, jsonBody = false) {
    const headers = { Accept: "application/json" };
    if (jsonBody) headers["Content-Type"] = "application/json";
    if (this.config.apiKey) headers.Authorization = `Bearer ${this.config.apiKey}`;
    const accountId = context.accountId ?? this.config.accountId;
    const userId = context.userId ?? this.config.userId;
    if (accountId) headers["X-OpenViking-Account"] = accountId;
    if (userId) headers["X-OpenViking-User"] = userId;
    return headers;
  }

  async request(method, path, { body, context } = {}) {
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), this.config.requestTimeoutMs);
    try {
      const response = await fetch(`${this.config.baseUrl}${path}`, {
        method,
        headers: this.headers(context, body !== undefined),
        body: body === undefined ? undefined : JSON.stringify(body),
        signal: controller.signal,
      });
      const text = await response.text();
      let payload = text;
      if (text) { try { payload = JSON.parse(text); } catch { /* text */ } }
      if (!response.ok) throw new Error(`OpenViking ${method} ${path} -> ${response.status}: ${typeof payload === "string" ? payload.slice(0, 600) : JSON.stringify(payload).slice(0, 600)}`);
      if (payload && typeof payload === "object" && "result" in payload) return payload.result;
      return payload;
    } finally { clearTimeout(timer); }
  }

  listSessions(context) { return this.request("GET", "/api/v1/sessions", { context }); }
  searchMemory(context, query, limit = 20) {
    return this.request("POST", "/api/v1/search/find", { context, body: { query, target_uri: "viking://~/memories", limit, read_content: true } });
  }
  listSkills(context) { return this.request("GET", "/api/v1/skills", { context }); }
  findSkills(context, query, limit = 20) { return this.request("POST", "/api/v1/skills/find", { context, body: { query, limit, level: [0, 1] } }); }
  getSkill(context, name, targetUri) {
    const q = new URLSearchParams({ level: "2", include_content: "true", include_files: "true" });
    if (targetUri) q.set("target_uri", targetUri);
    return this.request("GET", `/api/v1/skills/${encodeURIComponent(name)}?${q}`, { context });
  }
}
