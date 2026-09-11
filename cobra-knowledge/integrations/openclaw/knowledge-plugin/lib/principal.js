export function requireDurableProfileId(client) {
  const profileId = String(client?.authenticatedUserProfile?.profileId ?? "").trim();
  if (!profileId) throw new Error("leeclaw: a durable OpenClaw authenticated user profile is required");
  return profileId;
}

function requiredEnvironment(name) {
  const key = String(name ?? "").trim();
  const value = String(process.env[key] ?? "").trim();
  if (!key || !value) throw new Error(`leeclaw-knowledge: environment ${key || "<empty>"} is required`);
  return value;
}

export function knowledgePrincipal(client, workspaceRegistry, requestedWorkspaceId, minimumRole = "viewer") {
  const userId = requireDurableProfileId(client);
  const workspace = workspaceRegistry.requireRole(userId, requestedWorkspaceId, minimumRole);
  return Object.freeze({
    issuedBy: "openclaw",
    userId,
    externalUserId: userId,
    workspaceId: workspace.id,
    workspaceName: workspace.name,
    workspaceRole: workspace.role,
    tenantId: workspace.weknoraTenantId,
    weknoraApiKey: requiredEnvironment(workspace.weknoraApiKeyEnv),
    weknoraBaseUrl: workspace.weknoraBaseUrl,
    openvikingAccountId: workspace.openvikingAccountId,
  });
}
