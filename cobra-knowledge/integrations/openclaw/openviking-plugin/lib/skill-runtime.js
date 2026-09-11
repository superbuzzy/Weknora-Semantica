const TOOL_ALIASES = Object.freeze({
  "context.retrieve": "leeclaw_context_retrieve",
  "knowledge.retrieve": "leeclaw_context_retrieve",
  "context.get_evidence": "leeclaw_context_get_evidence",
  "knowledge.get_evidence": "leeclaw_context_get_evidence",
  Read: "read",
  Write: "write",
  Edit: "edit",
  Bash: "exec",
  WebSearch: "web_search",
  WebFetch: "web_fetch",
});

function list(value) {
  if (Array.isArray(value)) return value;
  for (const key of ["skills", "items", "results", "data"]) if (Array.isArray(value?.[key])) return value[key];
  return [];
}

function score(candidate) {
  const value = Number(candidate?.score ?? candidate?.similarity ?? 0);
  return Number.isFinite(value) ? value : 0;
}

function targetUri(candidate) {
  const root = String(candidate?.root_uri ?? candidate?.uri ?? "").replace(/\/+$/u, "");
  return root.includes("/") ? root.replace(/\/[^/]+$/u, "") : undefined;
}

function candidateSourceRank(candidate) {
  const root = String(candidate?.root_uri ?? candidate?.uri ?? "").trim();
  if (root.startsWith("viking://user/")) return 0;
  if (root.startsWith("viking://agent/skills/")) return 1;
  return 2;
}

function candidatePrecedes(a, b) {
  const sourceDiff = candidateSourceRank(a) - candidateSourceRank(b);
  if (sourceDiff !== 0) return sourceDiff;
  const scoreDiff = score(b) - score(a);
  if (scoreDiff !== 0) return scoreDiff;
  return String(a?.name ?? "").localeCompare(String(b?.name ?? ""));
}

function skillText(payload) {
  for (const value of [payload?.skill_md, payload?.skill_md_content, payload?.content, payload?.data?.skill_md, payload?.data?.content]) {
    if (typeof value === "string" && value.trim()) return value.trim();
  }
  return "";
}

function splitFrontmatter(text) {
  const match = String(text ?? "").match(/^---\s*\n([\s\S]*?)\n---\s*\n?([\s\S]*)$/u);
  return match ? { frontmatter: match[1], body: match[2].trim() } : { frontmatter: "", body: String(text ?? "").trim() };
}

function splitAllowedTools(text) {
  if (Array.isArray(text)) return text.map(String).map((x) => x.trim()).filter(Boolean);
  const source = String(text ?? "").trim().replace(/^\[|\]$/gu, "");
  const out = []; let current = ""; let depth = 0;
  for (const char of source) {
    if (char === "(") depth += 1;
    if (char === ")") depth -= 1;
    if (depth === 0 && (char === "," || /\s/u.test(char))) { if (current) { out.push(current); current = ""; } continue; }
    current += char;
  }
  if (current) out.push(current);
  return out.map((x) => x.replace(/^['"]|['"]$/gu, "").trim()).filter(Boolean);
}

function frontmatterAllowedTools(frontmatter) {
  const lines = String(frontmatter ?? "").split(/\r?\n/u);
  const index = lines.findIndex((value) => /^allowed[-_]tools\s*:/iu.test(value));
  if (index < 0) return { declared: false, tools: [] };

  const inline = lines[index].replace(/^allowed[-_]tools\s*:/iu, "").trim();
  if (inline) return { declared: true, tools: splitAllowedTools(inline) };

  const tools = [];
  for (let i = index + 1; i < lines.length; i += 1) {
    const line = lines[i];
    if (!/^\s+/u.test(line)) break;
    const item = line.match(/^\s*-\s*(.+?)\s*$/u);
    if (!item) continue;
    const value = item[1].replace(/^['"]|['"]$/gu, "").trim();
    if (value) tools.push(value);
  }
  return { declared: true, tools };
}

function normalizeAllowedTools(rawTools, toolAuthority) {
  toolAuthority?.assertActive?.();
  const result = [];
  for (const raw of rawTools) {
    const token = String(raw ?? "").trim();
    if (!token || token.includes("(")) continue;
    const mapped = TOOL_ALIASES[token] ?? token;
    if (toolAuthority?.allows && !toolAuthority.allows(mapped)) continue;
    if (!result.includes(mapped)) result.push(mapped);
  }
  return result;
}

export function parseResolvedSkill(candidate, payload, toolAuthority) {
  const rawText = skillText(payload);
  const parsed = splitFrontmatter(rawText);
  const fmTools = frontmatterAllowedTools(parsed.frontmatter);
  const payloadTools = payload?.allowed_tools ?? payload?.data?.allowed_tools;
  const explicitDeclared = payload?.allowed_tools_declared ?? payload?.data?.allowed_tools_declared;
  const declared = explicitDeclared !== undefined ? Boolean(explicitDeclared) : (fmTools.declared || (Array.isArray(payloadTools) && payloadTools.length > 0));
  const rawTools = Array.isArray(payloadTools) && payloadTools.length ? payloadTools : fmTools.tools;
  return {
    name: String(payload?.name ?? payload?.data?.name ?? candidate?.name ?? "").trim(),
    description: String(payload?.description ?? payload?.data?.description ?? candidate?.description ?? "").trim(),
    content: parsed.body,
    score: score(candidate),
    sourceUri: String(candidate?.root_uri ?? candidate?.uri ?? payload?.root_uri ?? "").trim(),
    matchReason: String(candidate?.match_reason ?? "").trim(),
    allowedToolsDeclared: declared,
    toolsAllow: declared ? normalizeAllowedTools(rawTools, toolAuthority) : undefined,
  };
}

export async function resolveSkillForTurn(client, principal, query, config, toolAuthority) {
  if (!config.skillRuntimeEnabled || !String(query ?? "").trim()) return null;
  const result = await client.findSkills(principal, String(query).trim(), config.skillResolveLimit, config.skillScoreThreshold);
  const candidates = list(result).filter((item) => score(item) >= config.skillScoreThreshold).sort(candidatePrecedes);
  const candidate = candidates[0];
  if (!candidate?.name) return null;
  const payload = await client.getSkill(principal, candidate.name, targetUri(candidate));
  const skill = parseResolvedSkill(candidate, payload, toolAuthority);
  return skill.content ? skill : null;
}

export function renderSkillContext(skill, maxChars = 20000) {
  if (!skill?.content) return undefined;
  const header = [
    "<leeclaw-active-skill>",
    `name: ${skill.name || "unnamed"}`,
    skill.description ? `description: ${skill.description}` : "",
    `source: ${skill.sourceUri || "OpenViking"}`,
    `match_score: ${skill.score}`,
    skill.matchReason ? `match_reason: ${skill.matchReason}` : "",
    "This Skill is the selected workspace-scoped operating procedure for this turn. Follow it unless it conflicts with host security, authorization, or tool policy. Enterprise facts must be retrieved through authoritative LeeClaw Knowledge tools when the Skill permits them; memory is not authoritative evidence.",
    "--- skill instructions ---",
  ].filter(Boolean).join("\n");
  const suffix = "\n</leeclaw-active-skill>";
  const room = Math.max(0, maxChars - header.length - suffix.length - 1);
  return `${header}\n${skill.content.slice(0, room)}${suffix}`;
}
