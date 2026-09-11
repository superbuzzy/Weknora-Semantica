import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { ContextRuntimeClient } from "../lib/context-runtime-client.js";
import { createKnowledgeRuntimeTool } from "../lib/agent-tools.js";
import { sessionKnowledgePrincipal } from "../lib/principal.js";
import { WorkspaceRegistry } from "../../workspace-core/index.js";

function setup() {
  const dir=fs.mkdtempSync(path.join(os.tmpdir(),"leeclaw-ctx-")); const rp=path.join(dir,"workspaces.json");
  fs.writeFileSync(rp,JSON.stringify({workspaces:[{id:"cq",members:[{profileId:"p1",role:"viewer"}],weknora:{tenantId:"42",apiKeyEnv:"WK_CTX",knowledgeBaseIds:["kb1","kb2"]},openviking:{accountId:"cq"}}]}));
  process.env.WK_CTX="wk-secret";
  const registry=new WorkspaceRegistry(rp,path.join(dir,"state.json"));
  const api={runtime:{agent:{session:{getSessionEntry(){return {createdActor:{type:"human",source:"profile",id:"p1"}};}}}}};
  const config={coreApiBaseUrl:"http://core.test",runtimeToken:"rt-secret",requestTimeoutMs:5000};
  return {registry,api,config,principal:sessionKnowledgePrincipal(api,registry,"s1","sid-ctx")};
}

async function withFetch(handler,fn){const old=globalThis.fetch;globalThis.fetch=handler;try{return await fn();}finally{globalThis.fetch=old;}}

test("Context Runtime service request carries only server-derived principal headers", async()=>{
  const {config,principal}=setup(); const client=new ContextRuntimeClient(config);
  await withFetch(async(url,init)=>{
    assert.equal(init.headers.Authorization,"Bearer rt-secret");
    assert.equal(init.headers["X-LeeClaw-Workspace-ID"],"cq");
    assert.equal(init.headers["X-LeeClaw-User-ID"],"p1");
    assert.equal(init.headers["X-LeeClaw-WeKnora-Tenant-ID"],"42");
    assert.equal(init.headers["X-LeeClaw-WeKnora-API-Key"],"wk-secret");
    const body=JSON.parse(init.body);
    assert.equal("workspaceId" in body,false); assert.equal("tenantId" in body,false);
    if(String(url).endsWith("/retrieve")) { assert.deepEqual(body.knowledge_base_ids,["kb1"]); return new Response(JSON.stringify({query:"q",knowledge:[]}),{status:200}); }
    assert.equal(String(url),"http://core.test/api/v1/runtime/context/evidence");
    assert.equal(body.chunk_id,"chunk-1"); assert.deepEqual(body.knowledge_base_ids,["kb1","kb2"]);
    return new Response(JSON.stringify({id:"chunk-1"}),{status:200});
  },async()=>{ await client.retrieve(principal,{query:"q"},["kb1"]); await client.getEvidence(principal,"chunk-1"); });
});

test("Agent tool schema cannot carry identity/workspace and rejects KB outside workspace", async()=>{
  const {registry,api}=setup();
  const contextClient={async retrieve(_p,_args,kbs){return {kbs};},async getEvidence(){return {id:"c1"};}};
  const tool=createKnowledgeRuntimeTool(api,registry,contextClient,{sessionKey:"s1",sessionId:"sid-tools"},"leeclaw_context_retrieve");
  assert.equal(tool.parameters.additionalProperties,false);
  for(const forbidden of ["workspaceId","tenantId","userId","accountId","apiKey"]) assert.equal(forbidden in tool.parameters.properties,false);
  const ok=await tool.execute("call",{query:"线路",knowledge_base_id:"kb2"}); assert.deepEqual(ok.structuredContent.kbs,["kb2"]);
  await assert.rejects(()=>tool.execute("call",{query:"线路",knowledge_base_id:"kb-x"}),/outside the active workspace scope/);
});


test("Evidence tool carries workspace Knowledge scope and exposes no scope override",async()=>{
  const {registry,api}=setup(); let seen;
  const contextClient={async getEvidence(p,chunk){seen={p,chunk};return {id:chunk};}};
  const tool=createKnowledgeRuntimeTool(api,registry,contextClient,{sessionKey:"s1",sessionId:"sid-evidence"},"leeclaw_context_get_evidence");
  assert.deepEqual(Object.keys(tool.parameters.properties),["chunk_id"]);
  const out=await tool.execute("call",{chunk_id:"c-1"});
  assert.equal(out.structuredContent.id,"c-1");
  assert.deepEqual(seen.p.knowledgeBaseIds,["kb1","kb2"]);
});

test("Tool factory disappears when session has no durable OpenClaw profile",()=>{
  const {registry}=setup(); const api={runtime:{agent:{session:{getSessionEntry(){return {createdActor:{type:"human",source:"channel",id:"x"}};}}}}};
  assert.equal(createKnowledgeRuntimeTool(api,registry,{}, {sessionKey:"s1",sessionId:"sid-missing"}, "leeclaw_context_retrieve"),null);
});
