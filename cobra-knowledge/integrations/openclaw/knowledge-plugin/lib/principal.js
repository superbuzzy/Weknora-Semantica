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

function fromWorkspace(userId, workspace) {
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
    knowledgeBaseIds: Object.freeze([...(workspace.weknoraKnowledgeBaseIds ?? [])]),
    openvikingAccountId: workspace.openvikingAccountId,
  });
}

export function knowledgePrincipal(client, workspaceRegistry, requestedWorkspaceId, minimumRole = "viewer") {
  const userId = requireDurableProfileId(client);
  return fromWorkspace(userId, workspaceRegistry.requireRole(userId, requestedWorkspaceId, minimumRole));
}

export function sessionKnowledgePrincipal(api, workspaceRegistry, sessionKey, sessionId) {
  const workspace = workspaceRegistry.resolveSession(api, sessionKey, sessionId);
  if (!workspace) return null;
  return fromWorkspace(workspace.profileId, workspace);
}

export function resolveKnowledgeBaseScope(principal, requestedKnowledgeBaseId) {
  const requested = String(requestedKnowledgeBaseId ?? "").trim();
  const allowed = Array.isArray(principal?.knowledgeBaseIds) ? principal.knowledgeBaseIds : [];
  if (!requested) return allowed;
  if (allowed.length > 0 && !allowed.includes(requested)) {
    throw new Error("leeclaw-knowledge: knowledge base is outside the active workspace scope");
  }
  return [requested];
}
