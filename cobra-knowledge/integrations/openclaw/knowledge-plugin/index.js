import { definePluginEntry } from "openclaw/plugin-sdk/plugin-entry";
import { resolveKnowledgeConfig } from "./lib/config.js";
import { knowledgePrincipal } from "./lib/principal.js";
import { WeKnoraClient } from "./lib/weknora-client.js";

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
      options.respond(false, { error: error instanceof Error ? error.message : String(error) });
    }
  }, { scope, profileAccess: "required" });
}

export default definePluginEntry({
  id: "leeclaw-knowledge",
  name: "LeeClaw Knowledge",
  description: "OpenClaw-native WeKnora knowledge workspace and ontology graph integration.",
  register(api) {
    const config = resolveKnowledgeConfig(api.pluginConfig ?? {});
    const client = new WeKnoraClient(config);
    const principal = (options) => knowledgePrincipal(options.client, config);

    api.session.controls.registerControlUiDescriptor({
      surface: "tab",
      id: "knowledge",
      label: "Knowledge",
      description: "Knowledge bases, documents, Wiki, graph, sharing and workspace access.",
      icon: "bookOpen",
      group: "control",
      requiredScopes: ["operator.read"],
    });

    registerMethod(api, "leeclaw.knowledge.health", "operator.read", async (options) => {
      const ctx = principal(options);
      await client.health(ctx);
      return {
        ok: true,
        workspaceId: ctx.workspaceId,
        userId: ctx.userId,
        weknora: config.weknoraBaseUrl,
        graph: config.graphApiBaseUrl || null,
      };
    });
    registerMethod(api, "leeclaw.knowledge.list", "operator.read", (options) =>
      client.listKnowledgeBases(principal(options), String(options.params.creator ?? "all")));
    registerMethod(api, "leeclaw.knowledge.get", "operator.read", (options) =>
      client.getKnowledgeBase(principal(options), requiredString(options.params, "id")));
    registerMethod(api, "leeclaw.knowledge.create", "operator.write", (options) =>
      client.createKnowledgeBase(principal(options), options.params));
    registerMethod(api, "leeclaw.knowledge.documents", "operator.read", (options) =>
      client.listDocuments(principal(options), requiredString(options.params, "kbId"), options.params));
    registerMethod(api, "leeclaw.knowledge.wiki", "operator.read", (options) =>
      client.listWikiPages(principal(options), requiredString(options.params, "kbId"), options.params));
    registerMethod(api, "leeclaw.knowledge.members", "operator.read", (options) =>
      client.listMembers(principal(options)));
    registerMethod(api, "leeclaw.knowledge.shares", "operator.read", (options) =>
      client.listShares(principal(options), requiredString(options.params, "kbId")));
    registerMethod(api, "leeclaw.knowledge.activity", "operator.read", (options) =>
      client.listActivity(principal(options), requiredString(options.params, "kbId")));
    registerMethod(api, "leeclaw.knowledge.graph", "operator.read", (options) =>
      client.graph(principal(options), requiredString(options.params, "kbId"), String(options.params.view ?? "entity")));
  },
});
