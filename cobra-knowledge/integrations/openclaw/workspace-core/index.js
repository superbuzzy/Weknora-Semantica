import fs from "node:fs";
import path from "node:path";

const ROLE_RANK = Object.freeze({ viewer: 1, editor: 2, admin: 3, owner: 4 });

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

function atomicWrite(file, value) {
  fs.mkdirSync(path.dirname(file), { recursive: true });
  const tmp = `${file}.${process.pid}.${Date.now()}.tmp`;
  fs.writeFileSync(tmp, `${JSON.stringify(value, null, 2)}\n`, { mode: 0o600 });
  fs.renameSync(tmp, file);
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
      const raw = JSON.parse(fs.readFileSync(this.statePath, "utf8"));
      return raw && typeof raw === "object" ? raw : {};
    } catch (error) {
      if (error?.code === "ENOENT") return {};
      throw error;
    }
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
      },
      openviking: {
        accountId: String(raw?.openviking?.accountId ?? id).trim() || id,
      },
    };
  }

  all() {
    return this.readRegistry().workspaces.map((item) => this.normalizeWorkspace(item));
  }

  memberFor(workspace, profileId) {
    return workspace.members.find((member) => member.profileId === profileId) ?? null;
  }

  listFor(profileId) {
    const state = this.readState();
    const selected = String(state[profileId] ?? "").trim();
    return this.all().flatMap((workspace) => {
      const member = this.memberFor(workspace, profileId);
      if (!member) return [];
      return [{
        id: workspace.id,
        name: workspace.name,
        description: workspace.description,
        role: member.role,
        selected: workspace.id === selected,
      }];
    });
  }

  resolve(profileId, requestedWorkspaceId) {
    const workspaces = this.all();
    const accessible = workspaces.filter((workspace) => this.memberFor(workspace, profileId));
    if (accessible.length === 0) throw new Error("leeclaw-workspace: authenticated profile has no workspace membership");
    const state = this.readState();
    const requested = String(requestedWorkspaceId ?? state[profileId] ?? "").trim();
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
      openvikingAccountId: workspace.openviking.accountId,
    });
  }

  switch(profileId, workspaceId) {
    const resolved = this.resolve(profileId, workspaceId);
    const state = this.readState();
    state[profileId] = resolved.id;
    atomicWrite(this.statePath, state);
    return resolved;
  }

  requireRole(profileId, workspaceId, minimumRole) {
    const resolved = this.resolve(profileId, workspaceId);
    const requiredRank = ROLE_RANK[normalizeRole(minimumRole)];
    if (ROLE_RANK[resolved.role] < requiredRank) {
      throw new Error(`leeclaw-workspace: ${minimumRole} role is required`);
    }
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
    const after = before.filter((item) => String(item.profileId) !== target);
    workspace.members = after;
    atomicWrite(this.registryPath, raw);
    return { removed: target };
  }
}
