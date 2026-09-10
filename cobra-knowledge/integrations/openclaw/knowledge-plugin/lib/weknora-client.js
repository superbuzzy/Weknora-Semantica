import { asArray, unwrap } from "./http.js";

function joinUrl(base, path) {
  return `${base}${path.startsWith("/") ? path : `/${path}`}`;
}

function encodeQuery(params = {}) {
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === "") continue;
    query.set(key, String(value));
  }
  const text = query.toString();
  return text ? `?${text}` : "";
}

function required(value, name) {
  const text = String(value ?? "").trim();
  if (!text) throw new Error(`leeclaw-knowledge: missing trusted ${name}`);
  return text;
}

export class WeKnoraClient {
  constructor(config) {
    this.config = config;
  }

  headers(context = {}, jsonBody = false) {
    const headers = { Accept: "application/json", "Accept-Language": "zh-CN" };
    if (jsonBody) headers["Content-Type"] = "application/json";
    headers["X-API-Key"] = required(this.config.weknoraApiKey, "WeKnora service credential");
    headers["X-Tenant-ID"] = required(context.tenantId, "WeKnora tenant");
    headers["X-External-User-ID"] = required(context.externalUserId, "OpenClaw principal");
    return headers;
  }

  async request(method, path, { body, context, baseUrl } = {}) {
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), this.config.requestTimeoutMs);
    try {
      const response = await fetch(joinUrl(baseUrl ?? this.config.weknoraBaseUrl, path), {
        method,
        headers: this.headers(context, body !== undefined),
        body: body === undefined ? undefined : JSON.stringify(body),
        signal: controller.signal,
      });
      const text = await response.text();
      let payload = text;
      if (text) {
        try { payload = JSON.parse(text); } catch { /* keep text */ }
      }
      if (!response.ok) {
        const detail = typeof payload === "string" ? payload : JSON.stringify(payload);
        throw new Error(`WeKnora ${method} ${path} -> ${response.status}: ${detail.slice(0, 600)}`);
      }
      return payload;
    } finally {
      clearTimeout(timer);
    }
  }

  health(context) { return this.request("GET", "/api/v1/knowledge-bases", { context }); }

  async listKnowledgeBases(context, creator = "all") {
    const payload = await this.request("GET", `/api/v1/knowledge-bases${encodeQuery({ creator: creator === "all" ? undefined : creator })}`, { context });
    return asArray(payload, ["knowledge_bases"]);
  }

  async getKnowledgeBase(context, id) {
    return unwrap(await this.request("GET", `/api/v1/knowledge-bases/${encodeURIComponent(id)}`, { context }));
  }

  async createKnowledgeBase(context, input) {
    return unwrap(await this.request("POST", "/api/v1/knowledge-bases", {
      context,
      body: {
        name: String(input.name ?? "").trim(),
        description: String(input.description ?? "").trim(),
        type: input.type === "faq" ? "faq" : "document",
        indexing_strategy: input.indexing_strategy ?? {
          vector_enabled: true,
          keyword_enabled: true,
          wiki_enabled: true,
          graph_enabled: true,
        },
      },
    }));
  }

  async listDocuments(context, kbId, params = {}) {
    const payload = await this.request("GET", `/api/v1/knowledge-bases/${encodeURIComponent(kbId)}/knowledge${encodeQuery({
      page: params.page ?? 1,
      page_size: params.page_size ?? 100,
      keyword: params.keyword,
      parse_status: params.parse_status,
      folder_path: params.folder_path,
    })}`, { context });
    return { raw: unwrap(payload), items: asArray(payload, ["knowledge", "items"]) };
  }

  async listWikiPages(context, kbId, params = {}) {
    const payload = await this.request("GET", `/api/v1/knowledgebase/${encodeURIComponent(kbId)}/wiki/pages${encodeQuery({
      page: params.page ?? 1,
      page_size: params.page_size ?? 100,
      query: params.query,
    })}`, { context });
    return { raw: unwrap(payload), items: asArray(payload, ["pages"]) };
  }

  async listMembers(context) {
    const tenantId = required(context.tenantId, "WeKnora tenant");
    const payload = await this.request("GET", `/api/v1/tenants/${encodeURIComponent(tenantId)}/members?page=1&page_size=100`, { context });
    return { raw: unwrap(payload), items: asArray(payload, ["members"]) };
  }

  async listShares(context, kbId) {
    const payload = await this.request("GET", `/api/v1/knowledge-bases/${encodeURIComponent(kbId)}/shares`, { context });
    return { raw: unwrap(payload), items: asArray(payload, ["shares"]) };
  }

  async listActivity(context, kbId) {
    const payload = await this.request("GET", `/api/v1/knowledge-bases/${encodeURIComponent(kbId)}/activity?limit=50`, { context });
    return { raw: unwrap(payload), items: asArray(payload, ["items", "logs", "activity"]) };
  }

  async graph(context, kbId, view = "entity") {
    if (!this.config.graphApiBaseUrl) throw new Error("leeclaw-knowledge: graphApiBaseUrl is not configured");
    return unwrap(await this.request("GET", `/api/v1/knowledge-bases/${encodeURIComponent(kbId)}/graph?view=${view === "ontology" ? "ontology" : "entity"}&limit=180`, {
      context,
      baseUrl: this.config.graphApiBaseUrl,
    }));
  }
}
