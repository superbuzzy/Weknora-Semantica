import { sessionKnowledgePrincipal, resolveKnowledgeBaseScope } from "./principal.js";

const RETRIEVE_PARAMETERS = {
  type: "object",
  additionalProperties: false,
  properties: {
    query: { type: "string", minLength: 1, description: "要检索的企业业务问题。" },
    knowledge_base_id: { type: "string", description: "可选。指定当前 Workspace 内的一个知识库；省略时使用 Workspace 默认知识范围。" },
    domain: { type: "string", description: "可选业务域。" },
    task: { type: "string", description: "可选任务类型，例如问答、问数、诊断、规划。" },
  },
  required: ["query"],
};

const EVIDENCE_PARAMETERS = {
  type: "object",
  additionalProperties: false,
  properties: { chunk_id: { type: "string", minLength: 1, description: "Context Pack 返回的 WeKnora chunk id。" } },
  required: ["chunk_id"],
};

function toolResult(value) {
  return { content: [{ type: "text", text: JSON.stringify(value, null, 2) }], structuredContent: value };
}

export function createKnowledgeRuntimeTool(api, workspaceRegistry, contextClient, toolContext, toolName) {
  const principal = sessionKnowledgePrincipal(api, workspaceRegistry, toolContext?.sessionKey, toolContext?.sessionId);
  if (!principal) return null;

  if (toolName === "leeclaw_context_retrieve") {
    return {
      name: toolName,
      label: "LeeClaw Knowledge Retrieve",
      description: "检索当前 Workspace 的企业知识、本体语义和可追溯证据。企业事实优先使用此工具，不把 Memory 当作权威事实。",
      parameters: RETRIEVE_PARAMETERS,
      async execute(_toolCallId, params) {
        const knowledgeBaseIds = resolveKnowledgeBaseScope(principal, params?.knowledge_base_id);
        return toolResult(await contextClient.retrieve(principal, params, knowledgeBaseIds));
      },
    };
  }

  if (toolName === "leeclaw_context_get_evidence") {
    return {
      name: toolName,
      label: "LeeClaw Evidence",
      description: "读取 Context Pack 中指定 chunk 的原始证据，用于核验和引用。",
      parameters: EVIDENCE_PARAMETERS,
      async execute(_toolCallId, params) {
        return toolResult(await contextClient.getEvidence(principal, params?.chunk_id));
      },
    };
  }
  return null;
}
