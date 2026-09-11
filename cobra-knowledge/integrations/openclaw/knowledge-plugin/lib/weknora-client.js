import { asArray, unwrap } from "./http.js";

function joinUrl(base, path) { return `${String(base).replace(/\/+$/u, "")}${path.startsWith("/") ? path : `/${path}`}`; }
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
function safeId(value, name = "id") { return encodeURIComponent(required(value, name)); }
function safeSlugPath(value) { return required(value, "slug").split("/").map(encodeURIComponent).join("/"); }

export class WeKnoraClient {
  constructor(config) { this.config = config; }

  headers(context = {}, mode = "json") {
    const headers = { Accept: "application/json", "Accept-Language": "zh-CN" };
    if (mode === "json") headers["Content-Type"] = "application/json";
    headers["X-API-Key"] = required(context.weknoraApiKey, "WeKnora workspace credential");
    headers["X-Tenant-ID"] = required(context.tenantId, "WeKnora tenant");
    headers["X-External-User-ID"] = required(context.externalUserId, "OpenClaw principal");
    return headers;
  }

  async request(method, path, { body, context, baseUrl, mode } = {}) {
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), this.config.requestTimeoutMs);
    const isForm = typeof FormData !== "undefined" && body instanceof FormData;
    try {
      const response = await fetch(joinUrl(baseUrl ?? context?.weknoraBaseUrl ?? this.config.weknoraBaseUrl, path), {
        method,
        headers: this.headers(context, isForm ? "form" : (mode ?? (body === undefined ? "none" : "json"))),
        body: body === undefined ? undefined : (isForm ? body : JSON.stringify(body)),
        signal: controller.signal,
      });
      const text = await response.text();
      let payload = text;
      if (text) { try { payload = JSON.parse(text); } catch { /* keep text */ } }
      if (!response.ok) {
        const detail = typeof payload === "string" ? payload : JSON.stringify(payload);
        throw new Error(`WeKnora ${method} ${path} -> ${response.status}: ${detail.slice(0, 1000)}`);
      }
      return payload;
    } finally { clearTimeout(timer); }
  }

  health(context) { return this.request("GET", "/api/v1/knowledge-bases", { context }); }

  async listKnowledgeBases(context, creator = "all") {
    const payload = await this.request("GET", `/api/v1/knowledge-bases${encodeQuery({ creator: creator === "all" ? undefined : creator })}`, { context });
    return asArray(payload, ["knowledge_bases"]);
  }
  async getKnowledgeBase(context, id) { return unwrap(await this.request("GET", `/api/v1/knowledge-bases/${safeId(id)}`, { context })); }
  async createKnowledgeBase(context, input) {
    const name = required(input?.name, "knowledge base name");
    return unwrap(await this.request("POST", "/api/v1/knowledge-bases", { context, body: {
      name,
      description: String(input?.description ?? "").trim(),
      type: input?.type === "faq" ? "faq" : "document",
      indexing_strategy: input?.indexing_strategy ?? { vector_enabled: true, keyword_enabled: true, wiki_enabled: true, graph_enabled: true },
    }}));
  }
  async updateKnowledgeBase(context, id, input) {
    return unwrap(await this.request("PUT", `/api/v1/knowledge-bases/${safeId(id)}`, { context, body: {
      name: required(input?.name, "knowledge base name"),
      description: String(input?.description ?? ""),
      ...(input?.config !== undefined ? { config: input.config } : {}),
    }}));
  }
  async deleteKnowledgeBase(context, id) { return unwrap(await this.request("DELETE", `/api/v1/knowledge-bases/${safeId(id)}`, { context })); }

  async listDocuments(context, kbId, params = {}) {
    const payload = await this.request("GET", `/api/v1/knowledge-bases/${safeId(kbId)}/knowledge${encodeQuery({
      page: params.page ?? 1, page_size: params.page_size ?? 100, keyword: params.keyword,
      parse_status: params.parse_status, folder_path: params.folder_path,
    })}`, { context });
    return { raw: unwrap(payload), items: asArray(payload, ["knowledge", "items"]) };
  }
  async uploadDocument(context, kbId, input) {
    const filename = required(input?.filename, "filename");
    const contentType = String(input?.contentType ?? "application/octet-stream");
    const bytes = Buffer.from(required(input?.base64, "base64 file content"), "base64");
    if (bytes.length > this.config.uploadMaxBytes) throw new Error(`leeclaw-knowledge: upload exceeds ${this.config.uploadMaxBytes} bytes`);
    const form = new FormData();
    form.append("file", new Blob([bytes], { type: contentType }), filename);
    if (input?.fileName) form.append("fileName", String(input.fileName));
    if (input?.metadata) form.append("metadata", JSON.stringify(input.metadata));
    if (input?.enableMultimodel !== undefined) form.append("enable_multimodel", String(Boolean(input.enableMultimodel)));
    if (Array.isArray(input?.tagIds) && input.tagIds.length) form.append("tag_ids", input.tagIds.join(","));
    if (input?.processConfig) form.append("process_config", JSON.stringify(input.processConfig));
    return unwrap(await this.request("POST", `/api/v1/knowledge-bases/${safeId(kbId)}/knowledge/file`, { context, body: form }));
  }
  async createKnowledgeFromURL(context, kbId, input) {
    return unwrap(await this.request("POST", `/api/v1/knowledge-bases/${safeId(kbId)}/knowledge/url`, { context, body: {
      url: required(input?.url, "url"),
      file_name: String(input?.fileName ?? ""), file_type: String(input?.fileType ?? ""),
      title: String(input?.title ?? ""), tag_ids: Array.isArray(input?.tagIds) ? input.tagIds : [],
      ...(input?.enableMultimodel !== undefined ? { enable_multimodel: Boolean(input.enableMultimodel) } : {}),
      ...(input?.processConfig ? { process_config: input.processConfig } : {}),
    }}));
  }
  async createManualKnowledge(context, kbId, input) {
    return unwrap(await this.request("POST", `/api/v1/knowledge-bases/${safeId(kbId)}/knowledge/manual`, { context, body: {
      title: required(input?.title, "title"), content: required(input?.content, "content"),
      status: String(input?.status ?? "published"), tag_ids: Array.isArray(input?.tagIds) ? input.tagIds : [],
      ...(input?.processConfig ? { process_config: input.processConfig } : {}),
    }}));
  }
  async deleteDocument(context, id) { return unwrap(await this.request("DELETE", `/api/v1/knowledge/${safeId(id)}`, { context })); }
  async reparseDocument(context, id, processConfig) { return unwrap(await this.request("POST", `/api/v1/knowledge/${safeId(id)}/reparse`, { context, body: processConfig ? { process_config: processConfig } : {} })); }
  async cancelDocumentParse(context, id) { return unwrap(await this.request("POST", `/api/v1/knowledge/${safeId(id)}/cancel-parse`, { context, body: {} })); }
  async listFolders(context, kbId) {
    const payload = await this.request("GET", `/api/v1/knowledge-bases/${safeId(kbId)}/knowledge/folders`, { context });
    return { raw: unwrap(payload), items: asArray(payload, ["folders", "items"]) };
  }

  async listTags(context, kbId, params = {}) {
    const payload = await this.request("GET", `/api/v1/knowledge-bases/${safeId(kbId)}/tags${encodeQuery({ page: params.page ?? 1, page_size: params.page_size ?? 100, keyword: params.keyword })}`, { context });
    return { raw: unwrap(payload), items: asArray(payload, ["tags", "items"]) };
  }
  async createTag(context, kbId, input) { return unwrap(await this.request("POST", `/api/v1/knowledge-bases/${safeId(kbId)}/tags`, { context, body: { name: required(input?.name, "tag name"), color: String(input?.color ?? ""), sort_order: Number(input?.sortOrder ?? 0) } })); }
  async updateTag(context, kbId, tagId, input) { return unwrap(await this.request("PUT", `/api/v1/knowledge-bases/${safeId(kbId)}/tags/${safeId(tagId, "tagId")}`, { context, body: input })); }
  async deleteTag(context, kbId, tagId, force = false) { return unwrap(await this.request("DELETE", `/api/v1/knowledge-bases/${safeId(kbId)}/tags/${safeId(tagId, "tagId")}${encodeQuery({ force: force ? "true" : undefined })}`, { context, body: {} })); }

  async listFAQ(context, kbId, params = {}) {
    const payload = await this.request("GET", `/api/v1/knowledge-bases/${safeId(kbId)}/faq/entries${encodeQuery({ page: params.page ?? 1, page_size: params.page_size ?? 100, keyword: params.keyword })}`, { context });
    return { raw: unwrap(payload), items: asArray(payload, ["entries", "items"]) };
  }
  async createFAQ(context, kbId, input) { return unwrap(await this.request("POST", `/api/v1/knowledge-bases/${safeId(kbId)}/faq/entry`, { context, body: input })); }
  async updateFAQ(context, kbId, entryId, input) { return unwrap(await this.request("PUT", `/api/v1/knowledge-bases/${safeId(kbId)}/faq/entries/${safeId(entryId, "entryId")}`, { context, body: input })); }
  async deleteFAQ(context, kbId, ids) { return unwrap(await this.request("DELETE", `/api/v1/knowledge-bases/${safeId(kbId)}/faq/entries`, { context, body: { ids: (ids ?? []).map(Number).filter(Number.isFinite) } })); }

  async listWikiPages(context, kbId, params = {}) {
    const payload = await this.request("GET", `/api/v1/knowledgebase/${safeId(kbId)}/wiki/pages${encodeQuery({ page: params.page ?? 1, page_size: params.page_size ?? 100, query: params.query })}`, { context });
    return { raw: unwrap(payload), items: asArray(payload, ["pages"]) };
  }
  async createWikiPage(context, kbId, input) { return unwrap(await this.request("POST", `/api/v1/knowledgebase/${safeId(kbId)}/wiki/pages`, { context, body: input })); }
  async updateWikiPage(context, kbId, slug, input) { return unwrap(await this.request("PUT", `/api/v1/knowledgebase/${safeId(kbId)}/wiki/pages/${safeSlugPath(slug)}`, { context, body: input })); }
  async deleteWikiPage(context, kbId, slug) { return unwrap(await this.request("DELETE", `/api/v1/knowledgebase/${safeId(kbId)}/wiki/pages/${safeSlugPath(slug)}`, { context })); }

  async hybridSearch(context, kbId, input) { return unwrap(await this.request("POST", `/api/v1/knowledge-bases/${safeId(kbId)}/hybrid-search`, { context, body: input })); }
  async listOrganizations(context) {
    const payload = await this.request("GET", "/api/v1/organizations", { context });
    return { raw: unwrap(payload), items: asArray(payload, ["organizations", "items"]) };
  }
  async listShares(context, kbId) {
    const payload = await this.request("GET", `/api/v1/knowledge-bases/${safeId(kbId)}/shares`, { context });
    return { raw: unwrap(payload), items: asArray(payload, ["shares"]) };
  }
  async createShare(context, kbId, organizationId, permission) { return unwrap(await this.request("POST", `/api/v1/knowledge-bases/${safeId(kbId)}/shares`, { context, body: { organization_id: required(organizationId, "organizationId"), permission: String(permission ?? "viewer") } })); }
  async updateShare(context, kbId, shareId, permission) { return unwrap(await this.request("PUT", `/api/v1/knowledge-bases/${safeId(kbId)}/shares/${safeId(shareId, "shareId")}`, { context, body: { permission: String(permission ?? "viewer") } })); }
  async deleteShare(context, kbId, shareId) { return unwrap(await this.request("DELETE", `/api/v1/knowledge-bases/${safeId(kbId)}/shares/${safeId(shareId, "shareId")}`, { context })); }
  async graph(context, kbId, view = "entity") {
    if (!this.config.graphApiBaseUrl) throw new Error("leeclaw-knowledge: graphApiBaseUrl is not configured");
    return unwrap(await this.request("GET", `/api/v1/knowledge-bases/${safeId(kbId)}/graph?view=${view === "ontology" ? "ontology" : "entity"}&limit=240`, { context, baseUrl: this.config.graphApiBaseUrl }));
  }
}
