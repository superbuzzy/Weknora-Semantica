import { definePluginEntry } from "openclaw/plugin-sdk/plugin-entry";
import { OpenVikingClient, resolveOpenVikingConfig } from "./lib/client.js";

function registerMethod(api, name, scope, handler) {
  api.registerGatewayMethod(name, async (options) => {
    try { options.respond(true, await handler(options)); }
    catch (error) { options.respond(false, { error: error instanceof Error ? error.message : String(error) }); }
  }, { scope });
}

export default definePluginEntry({
  id: "leeclaw-openviking",
  name: "LeeClaw Memory & Skills",
  description: "OpenClaw-native OpenViking Memory and Skill console.",
  register(api) {
    const config = resolveOpenVikingConfig(api.pluginConfig ?? {});
    const client = new OpenVikingClient(config);
    const identity = (options) => ({
      accountId: options.params.accountId ?? config.accountId,
      userId: options.params.userId ?? (config.forwardAuthenticatedUser ? options.client?.authenticatedUserId : undefined) ?? config.userId,
    });

    api.session.controls.registerControlUiDescriptor({ surface: "tab", id: "memory", label: "Memory", icon: "brain", group: "control", requiredScopes: ["operator.read"] });
    api.session.controls.registerControlUiDescriptor({ surface: "tab", id: "skills", label: "Skills", icon: "puzzle", group: "control", requiredScopes: ["operator.read"] });

    registerMethod(api, "leeclaw.memory.sessions", "operator.read", (options) => client.listSessions(identity(options)));
    registerMethod(api, "leeclaw.memory.search", "operator.read", (options) => client.searchMemory(identity(options), String(options.params.query ?? ""), Number(options.params.limit ?? 20)));
    registerMethod(api, "leeclaw.skills.list", "operator.read", (options) => client.listSkills(identity(options)));
    registerMethod(api, "leeclaw.skills.find", "operator.read", (options) => client.findSkills(identity(options), String(options.params.query ?? ""), Number(options.params.limit ?? 20)));
    registerMethod(api, "leeclaw.skills.get", "operator.read", (options) => client.getSkill(identity(options), String(options.params.name ?? ""), options.params.targetUri ? String(options.params.targetUri) : undefined));
  },
});
