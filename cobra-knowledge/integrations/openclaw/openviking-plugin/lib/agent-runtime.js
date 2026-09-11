import { sessionPrincipal } from "./principal.js";
import { recallMemoryContext, registerMemoryCaptureRuntime } from "./memory-runtime.js";
import { renderSkillContext, resolveSkillForTurn } from "./skill-runtime.js";

export function registerLeeClawAgentRuntime(api, config, client, workspaceRegistry) {
  registerMemoryCaptureRuntime(api, config, client, workspaceRegistry);
  if (!config.memoryRuntimeEnabled && !config.skillRuntimeEnabled) return;

  api.on("before_prompt_build", async (event, ctx) => {
    const principal = sessionPrincipal(api, workspaceRegistry, ctx?.sessionKey, ctx?.sessionId);
    if (!principal) return undefined;
    const query = String(event?.prompt ?? "").trim();
    if (!query) return undefined;

    let memoryContext;
    let skill;
    if (config.memoryRuntimeEnabled) {
      try { memoryContext = await recallMemoryContext(client, principal, query, config); }
      catch (error) { api.logger.warn(`leeclaw-openviking: memory recall skipped: ${error instanceof Error ? error.message : String(error)}`); }
    }
    if (config.skillRuntimeEnabled) {
      try { skill = await resolveSkillForTurn(client, principal, query, config, ctx?.toolAuthority); }
      catch (error) { api.logger.warn(`leeclaw-openviking: skill resolve skipped: ${error instanceof Error ? error.message : String(error)}`); }
    }

    const parts = [memoryContext, renderSkillContext(skill, config.skillMaxContentChars)].filter(Boolean);
    if (!parts.length && !skill?.allowedToolsDeclared) return undefined;
    return {
      ...(parts.length ? { prependContext: parts.join("\n\n") } : {}),
      ...(skill?.allowedToolsDeclared ? { toolsAllow: skill.toolsAllow ?? [] } : {}),
    };
  }, { requiresToolAuthority: true, priority: 50, registrationId: "leeclaw-runtime-context" });
}
