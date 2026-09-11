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

export function sessionPrincipal(_api, workspaceRegistry, sessionKey, sessionId) {
  const workspace = workspaceRegistry.resolveSession(_api, sessionKey, sessionId);
  return workspace ? fromWorkspace(workspace.profileId, workspace) : null;
}
