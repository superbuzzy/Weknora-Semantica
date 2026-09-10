import { definePluginEntry } from "openclaw/plugin-sdk/plugin-entry";
import { OpenVikingClient, resolveOpenVikingConfig } from "./lib/client.js";
import { registerIdentityAwareMemoryRuntime } from "./lib/memory-runtime.js";
import { openVikingPrincipal } from "./lib/principal.js";

function registerMethod(api, name, scope, handler) {
  api.registerGatewayMethod(name, async (options) => {
    try { options.respond(true, await handler(options)); }
    catch (error) { options.respond(false, { error: error instanceof Error ? error.message : String(error) }); }
  }, { scope, profileAccess: "required" });
}

export default definePluginEntry({
  id: "leeclaw-openviking",
  name: "LeeClaw Memory & Skills",
  description: "OpenClaw-native OpenViking Memory and Skill integration using the durable OpenClaw profile as the only user identity.",
  register(api) {
    const config = resolveOpenVikingConfig(api.pluginConfig ?? {});
    const client = new OpenVikingClient(config);
    const principal = (options) => openVikingPrincipal(options.client, config);

    api.session.controls.registerControlUiDescriptor({ surface: "tab", id: "memory", label: "Memory", icon: "brain", group: "control", requiredScopes: ["operator.read"] });
    api.session.controls.registerControlUiDescriptor({ surface: "tab", id: "skills", label: "Skills", icon: "puzzle", group: "control", requiredScopes: ["operator.read"] });

    registerMethod(api, "leeclaw.memory.sessions", "operator.read", (options) => client.listSessions(principal(options)));
    registerMethod(api, "leeclaw.memory.search", "operator.read", (options) => client.searchMemory(principal(options), String(options.params.query ?? ""), Number(options.params.limit ?? 20)));
    registerMethod(api, "leeclaw.skills.list", "operator.read", (options) => client.listSkills(principal(options)));
    registerMethod(api, "leeclaw.skills.find", "operator.read", (options) => client.findSkills(principal(options), String(options.params.query ?? ""), Number(options.params.limit ?? 20)));
    registerMethod(api, "leeclaw.skills.get", "operator.read", (options) => client.getSkill(principal(options), String(options.params.name ?? ""), options.params.targetUri ? String(options.params.targetUri) : undefined));

    registerIdentityAwareMemoryRuntime(api, config, client);
  },
});
