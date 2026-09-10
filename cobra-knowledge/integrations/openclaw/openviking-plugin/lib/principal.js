export function requireDurableProfileId(client) {
  const profileId = String(client?.authenticatedUserProfile?.profileId ?? "").trim();
  if (!profileId) throw new Error("leeclaw: a durable OpenClaw authenticated user profile is required");
  return profileId;
}

export function openVikingPrincipal(client, config) {
  const userId = requireDurableProfileId(client);
  return Object.freeze({ issuedBy: "openclaw", workspaceId: config.workspaceId, accountId: config.workspaceId, userId });
}

export function sessionPrincipal(api, config, sessionKey) {
  const key = String(sessionKey ?? "").trim();
  if (!key) return null;
  const entry = api.runtime.agent.session.getSessionEntry({ sessionKey: key, readConsistency: "latest" });
  const actor = entry?.createdActor;
  if (actor?.type !== "human" || actor?.source !== "profile" || !String(actor.id ?? "").trim()) return null;
  const userId = String(actor.id).trim();
  return Object.freeze({ issuedBy: "openclaw-session", workspaceId: config.workspaceId, accountId: config.workspaceId, userId });
}
