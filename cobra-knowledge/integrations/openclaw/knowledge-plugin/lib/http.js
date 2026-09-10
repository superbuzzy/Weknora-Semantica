export function unwrap(payload) {
  if (payload && typeof payload === "object") {
    if (Object.prototype.hasOwnProperty.call(payload, "result")) return payload.result;
    if (Object.prototype.hasOwnProperty.call(payload, "data")) return payload.data;
  }
  return payload;
}

export function asArray(value, keys = []) {
  const root = unwrap(value);
  if (Array.isArray(root)) return root;
  if (!root || typeof root !== "object") return [];
  for (const key of keys) {
    if (Array.isArray(root[key])) return root[key];
  }
  for (const key of ["items", "list", "records", "rows", "knowledge_bases", "knowledge", "pages", "members", "shares"]) {
    if (Array.isArray(root[key])) return root[key];
  }
  return [];
}

export function readTextField(value, keys, fallback = "") {
  if (!value || typeof value !== "object") return fallback;
  for (const key of keys) {
    const item = value[key];
    if (item !== undefined && item !== null && String(item).trim()) return String(item);
  }
  return fallback;
}

export function safeJson(value) {
  try {
    return JSON.stringify(value);
  } catch {
    return String(value);
  }
}
