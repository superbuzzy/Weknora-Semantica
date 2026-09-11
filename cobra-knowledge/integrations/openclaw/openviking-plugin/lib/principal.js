export function requireDurableProfileId(client) {
  const profileId = String(client?.authenticatedUserProfile?.profileId ?? "").trim();
  if (!profileId) throw new Error("leeclaw: a durable OpenClaw authenticated user profile is required");
  return profileId;
}

function fromWorkspace(profileId, workspace) {
  return Object.freeze({ issuedBy: "openclaw", workspaceId: workspace.id, workspaceName: workspace.name, workspaceRole: workspace.role, accountId: workspace.openvikingAccountId, userId: profileId });
}

export function openVikingPrincipal(client, workspaceRegistry, requestedWorkspaceId) {
  const profileId = requireDurableProfileId(client);
  return fromWorkspace(profileId, workspaceRegistry.resolve(profileId, requestedWorkspaceId));
}

export function sessionPrincipal(api, workspaceRegistry, sessionKey) {
  const key = String(sessionKey ?? "").trim();
  if (!key) return null;
  const entry = api.runtime.agent.session.getSessionEntry({ sessionKey: key, readConsistency: "latest" });
  const actor = entry?.createdActor;
  if (actor?.type !== "human" || actor?.source !== "profile" || !String(actor.id ?? "").trim()) return null;
  const profileId = String(actor.id).trim();
  try { return fromWorkspace(profileId, workspaceRegistry.resolve(profileId)); }
  catch { return null; }
}
