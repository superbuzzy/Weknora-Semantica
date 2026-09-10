export function requireDurableProfileId(client) {
  const profileId = String(client?.authenticatedUserProfile?.profileId ?? "").trim();
  if (!profileId) {
    throw new Error("leeclaw: a durable OpenClaw authenticated user profile is required");
  }
  return profileId;
}

export function knowledgePrincipal(client, config) {
  const userId = requireDurableProfileId(client);
  return Object.freeze({
    issuedBy: "openclaw",
    userId,
    externalUserId: userId,
    workspaceId: config.workspaceId,
    tenantId: config.weknoraTenantId,
  });
}
