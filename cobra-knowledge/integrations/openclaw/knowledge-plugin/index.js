import { definePluginEntry } from "openclaw/plugin-sdk/plugin-entry";
import { resolveKnowledgeConfig } from "./lib/config.js";
import { WeKnoraClient } from "./lib/weknora-client.js";

function principalFromRequest(options, config, params = {}) {
  const authenticated = options.client?.authenticatedUserId;
  return {
    tenantId: params.tenantId ?? config.tenantId,
    externalUserId: params.externalUserId ?? (config.forwardAuthenticatedUser ? authenticated : undefined) ?? config.externalUserId,
  };
}

function requiredString(params, key) {
  const value = String(params?.[key] ?? "").trim();
  if (!value) throw new Error(`${key} is required`);
  return value;
}

function registerMethod(api, name, scope, handler) {
  api.registerGatewayMethod(name, async (options) => {
    try {
      const result = await handler(options);
      options.respond(true, result);
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      options.respond(false, { error: message });
    }
  }, { scope });
}

export default definePluginEntry({
  id: "leeclaw-knowledge",
  name: "LeeClaw Knowledge",
  description: "OpenClaw-native WeKnora knowledge workspace and ontology graph integration.",
  register(api) {
    const config = resolveKnowledgeConfig(api.pluginConfig ?? {});
    const client = new WeKnoraClient(config);

    api.session.controls.registerControlUiDescriptor({
      surface: "tab",
      id: "knowledge",
      label: "Knowledge",
      description: "Knowledge bases, documents, Wiki, graph, sharing and workspace access.",
      icon: "bookOpen",
      group: "control",
      requiredScopes: ["operator.read"],
    });

    const ctx = (options) => principalFromRequest(options, config, options.params);

    registerMethod(api, "leeclaw.knowledge.health", "operator.read", async (options) => {
      await client.health(ctx(options));
      return { ok: true, weknora: config.weknoraBaseUrl, graph: config.graphApiBaseUrl || null };
    });
    registerMethod(api, "leeclaw.knowledge.list", "operator.read", (options) =>
      client.listKnowledgeBases(ctx(options), String(options.params.creator ?? "all")));
    registerMethod(api, "leeclaw.knowledge.get", "operator.read", (options) =>
      client.getKnowledgeBase(ctx(options), requiredString(options.params, "id")));
    registerMethod(api, "leeclaw.knowledge.create", "operator.write", (options) =>
      client.createKnowledgeBase(ctx(options), options.params));
    registerMethod(api, "leeclaw.knowledge.documents", "operator.read", (options) =>
      client.listDocuments(ctx(options), requiredString(options.params, "kbId"), options.params));
    registerMethod(api, "leeclaw.knowledge.wiki", "operator.read", (options) =>
      client.listWikiPages(ctx(options), requiredString(options.params, "kbId"), options.params));
    registerMethod(api, "leeclaw.knowledge.members", "operator.read", (options) => {
      const tenantId = options.params.tenantId ?? config.tenantId;
      if (!tenantId) throw new Error("tenantId is required for workspace members");
      return client.listMembers(ctx(options), tenantId);
    });
    registerMethod(api, "leeclaw.knowledge.shares", "operator.read", (options) =>
      client.listShares(ctx(options), requiredString(options.params, "kbId")));
    registerMethod(api, "leeclaw.knowledge.activity", "operator.read", (options) =>
      client.listActivity(ctx(options), requiredString(options.params, "kbId")));
    registerMethod(api, "leeclaw.knowledge.graph", "operator.read", (options) =>
      client.graph(ctx(options), requiredString(options.params, "kbId"), String(options.params.view ?? "entity")));
  },
});
