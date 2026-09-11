import { createHash } from "node:crypto";
import fs from "node:fs";
import path from "node:path";

const ROLE_RANK = Object.freeze({ viewer: 1, editor: 2, admin: 3, owner: 4 });
const STATE_VERSION = 2;

function required(value, name) {
  const text = String(value ?? "").trim();
  if (!text) throw new Error(`leeclaw-workspace: ${name} is required`);
  return text;
}

function normalizeRole(value) {
  const role = String(value ?? "viewer").trim().toLowerCase();
  if (!(role in ROLE_RANK)) throw new Error(`leeclaw-workspace: invalid role ${role}`);
  return role;
}

function normalizeStringArray(value) {
  if (!Array.isArray(value)) return [];
  return [...new Set(value.map((item) => String(item ?? "").trim()).filter(Boolean))];
}

function atomicWrite(file, value) {
  fs.mkdirSync(path.dirname(file), { recursive: true });
  const tmp = `${file}.${process.pid}.${Date.now()}.tmp`;
  fs.writeFileSync(tmp, `${JSON.stringify(value, null, 2)}\n`, { mode: 0o600 });
  fs.renameSync(tmp, file);
}

function emptyState() {
  return { version: STATE_VERSION, profiles: {}, sessions: {} };
}

function normalizeState(raw) {
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) return emptyState();
  if (raw.version === STATE_VERSION) {
    return {
      version: STATE_VERSION,
      profiles: raw.profiles && typeof raw.profiles === "object" && !Array.isArray(raw.profiles) ? { ...raw.profiles } : {},
      sessions: raw.sessions && typeof raw.sessions === "object" && !Array.isArray(raw.sessions) ? { ...raw.sessions } : {},
    };
  }

  // v0.7 state was a flat profileId -> workspaceId map. Upgrade it once into
  // the canonical v0.8 state shape; session bindings are intentionally empty.
  const profiles = {};
  for (const [profileId, workspaceId] of Object.entries(raw)) {
    const profile = String(profileId ?? "").trim();
    const workspace = String(workspaceId ?? "").trim();
    if (profile && workspace && typeof workspaceId === "string") profiles[profile] = workspace;
  }
  return { version: STATE_VERSION, profiles, sessions: {} };
}

export function sessionProfileId(api, sessionKey) {
  const key = String(sessionKey ?? "").trim();
  if (!key) return null;
  const entry = api?.runtime?.agent?.session?.getSessionEntry?.({ sessionKey: key, readConsistency: "latest" });
  const actor = entry?.createdActor;
  if (actor?.type !== "human" || actor?.source !== "profile") return null;
  const profileId = String(actor.id ?? "").trim();
  return profileId || null;
}

export function sessionBindingKey(sessionId, sessionKey) {
  const stable = String(sessionId ?? "").trim();
  const fallback = String(sessionKey ?? "").trim();
  if (!stable && !fallback) return null;
  const source = stable ? `session-id\0${stable}` : `session-key\0${fallback}`;
  return createHash("sha256").update(source).digest("hex");
}

export class WorkspaceRegistry {
  constructor(registryPath, statePath) {
    this.registryPath = required(registryPath, "workspaceRegistryPath");
    this.statePath = required(statePath ?? `${this.registryPath}.state.json`, "workspaceStatePath");
  }

  readRegistry() {
    const raw = JSON.parse(fs.readFileSync(this.registryPath, "utf8"));
    if (!Array.isArray(raw.workspaces) || raw.workspaces.length === 0) {
      throw new Error("leeclaw-workspace: workspaces must be a non-empty array");
    }
    return raw;
  }

  readState() {
    try {
      return normalizeState(JSON.parse(fs.readFileSync(this.statePath, "utf8")));
    } catch (error) {
      if (error?.code === "ENOENT") return emptyState();
      throw error;
    }
  }

  writeState(state) {
    atomicWrite(this.statePath, normalizeState(state));
  }

  normalizeWorkspace(raw) {
    const id = required(raw?.id, "workspace id");
    const members = Array.isArray(raw?.members) ? raw.members.map((member) => ({
      profileId: required(member?.profileId, `member profileId for ${id}`),
      role: normalizeRole(member?.role),
      displayName: String(member?.displayName ?? "").trim() || undefined,
    })) : [];
    return {
      id,
      name: String(raw?.name ?? id).trim() || id,
      description: String(raw?.description ?? "").trim(),
      members,
      weknora: {
        tenantId: required(raw?.weknora?.tenantId, `weknora.tenantId for ${id}`),
        apiKeyEnv: required(raw?.weknora?.apiKeyEnv, `weknora.apiKeyEnv for ${id}`),
        baseUrl: String(raw?.weknora?.baseUrl ?? "").trim() || undefined,
        knowledgeBaseIds: normalizeStringArray(raw?.weknora?.knowledgeBaseIds),
      },
      openviking: {
        accountId: String(raw?.openviking?.accountId ?? id).trim() || id,
      },
    };
  }

  all() { return this.readRegistry().workspaces.map((item) => this.normalizeWorkspace(item)); }
  memberFor(workspace, profileId) { return workspace.members.find((member) => member.profileId === profileId) ?? null; }

  listFor(profileId) {
    const state = this.readState();
    const selected = String(state.profiles[profileId] ?? "").trim();
    return this.all().flatMap((workspace) => {
      const member = this.memberFor(workspace, profileId);
      if (!member) return [];
      return [{ id: workspace.id, name: workspace.name, description: workspace.description, role: member.role, selected: workspace.id === selected }];
    });
  }

  resolve(profileId, requestedWorkspaceId) {
    const workspaces = this.all();
    const accessible = workspaces.filter((workspace) => this.memberFor(workspace, profileId));
    if (accessible.length === 0) throw new Error("leeclaw-workspace: authenticated profile has no workspace membership");
    const state = this.readState();
    const requested = String(requestedWorkspaceId ?? state.profiles[profileId] ?? "").trim();
    const workspace = (requested ? accessible.find((item) => item.id === requested) : null) ?? accessible[0];
    if (requested && workspace.id !== requested) throw new Error("leeclaw-workspace: requested workspace is not accessible to this profile");
    const member = this.memberFor(workspace, profileId);
    return Object.freeze({
      id: workspace.id,
      name: workspace.name,
      description: workspace.description,
      role: member.role,
      profileId,
      weknoraTenantId: workspace.weknora.tenantId,
      weknoraApiKeyEnv: workspace.weknora.apiKeyEnv,
      weknoraBaseUrl: workspace.weknora.baseUrl,
      weknoraKnowledgeBaseIds: Object.freeze([...workspace.weknora.knowledgeBaseIds]),
      openvikingAccountId: workspace.openviking.accountId,
    });
  }

  resolveSession(api, sessionKey, sessionId) {
    const profileId = sessionProfileId(api, sessionKey);
    const bindingKey = sessionBindingKey(sessionId, sessionKey);
    if (!profileId || !bindingKey) return null;

    const state = this.readState();
    const binding = state.sessions[bindingKey];
    if (binding) {
      if (String(binding.profileId ?? "").trim() !== profileId) return null;
      const workspaceId = String(binding.workspaceId ?? "").trim();
      if (!workspaceId) return null;
      try { return this.resolve(profileId, workspaceId); } catch { return null; }
    }

    let workspace;
    try { workspace = this.resolve(profileId); } catch { return null; }
    state.sessions[bindingKey] = {
      profileId,
      workspaceId: workspace.id,
      boundAt: new Date().toISOString(),
    };
    this.writeState(state);
    return workspace;
  }

  switch(profileId, workspaceId) {
    const resolved = this.resolve(profileId, workspaceId);
    const state = this.readState();
    state.profiles[profileId] = resolved.id;
    this.writeState(state);
    return resolved;
  }

  requireRole(profileId, workspaceId, minimumRole) {
    const resolved = this.resolve(profileId, workspaceId);
    const requiredRank = ROLE_RANK[normalizeRole(minimumRole)];
    if (ROLE_RANK[resolved.role] < requiredRank) throw new Error(`leeclaw-workspace: ${minimumRole} role is required`);
    return resolved;
  }

  listMembers(profileId, workspaceId) {
    const resolved = this.requireRole(profileId, workspaceId, "viewer");
    const workspace = this.all().find((item) => item.id === resolved.id);
    return workspace.members.map((member) => ({ ...member }));
  }

  upsertMember(callerProfileId, workspaceId, targetProfileId, role, displayName) {
    this.requireRole(callerProfileId, workspaceId, "owner");
    const raw = this.readRegistry();
    const index = raw.workspaces.findIndex((item) => String(item.id) === String(workspaceId));
    if (index < 0) throw new Error("leeclaw-workspace: workspace not found");
    raw.workspaces[index].members = Array.isArray(raw.workspaces[index].members) ? raw.workspaces[index].members : [];
    const members = raw.workspaces[index].members;
    const target = required(targetProfileId, "targetProfileId");
    const normalized = { profileId: target, role: normalizeRole(role), ...(String(displayName ?? "").trim() ? { displayName: String(displayName).trim() } : {}) };
    const memberIndex = members.findIndex((item) => String(item.profileId) === target);
    if (memberIndex >= 0 && normalizeRole(members[memberIndex].role) === "owner" && normalized.role !== "owner") {
      const ownerCount = members.filter((item) => normalizeRole(item.role) === "owner").length;
      if (ownerCount <= 1) throw new Error("leeclaw-workspace: cannot demote the last owner");
    }
    if (memberIndex >= 0) members[memberIndex] = normalized; else members.push(normalized);
    atomicWrite(this.registryPath, raw);
    return normalized;
  }

  removeMember(callerProfileId, workspaceId, targetProfileId) {
    this.requireRole(callerProfileId, workspaceId, "owner");
    if (callerProfileId === targetProfileId) throw new Error("leeclaw-workspace: owner cannot remove self from active workspace");
    const raw = this.readRegistry();
    const workspace = raw.workspaces.find((item) => String(item.id) === String(workspaceId));
    if (!workspace) throw new Error("leeclaw-workspace: workspace not found");
    const before = Array.isArray(workspace.members) ? workspace.members : [];
    const target = required(targetProfileId, "targetProfileId");
    const targetMember = before.find((item) => String(item.profileId) === target);
    if (!targetMember) throw new Error("leeclaw-workspace: member not found");
    if (normalizeRole(targetMember.role) === "owner") {
      const ownerCount = before.filter((item) => normalizeRole(item.role) === "owner").length;
      if (ownerCount <= 1) throw new Error("leeclaw-workspace: cannot remove the last owner");
    }
    workspace.members = before.filter((item) => String(item.profileId) !== target);
    atomicWrite(this.registryPath, raw);
    return { removed: target };
  }
}
