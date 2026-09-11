import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { resolveKnowledgeConfig } from "../lib/config.js";
import { knowledgePrincipal, requireDurableProfileId } from "../lib/principal.js";
import { WeKnoraClient } from "../lib/weknora-client.js";
import { WorkspaceRegistry } from "../lib/workspace-registry.js";

function jsonResponse(body, status = 200) { return new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } }); }
async function withFetch(handler, fn) { const original=globalThis.fetch; globalThis.fetch=handler; try{await fn();}finally{globalThis.fetch=original;} }
function setup() {
  const dir=fs.mkdtempSync(path.join(os.tmpdir(),"leeclaw-k-")); const rp=path.join(dir,"workspaces.json");
  fs.writeFileSync(rp,JSON.stringify({workspaces:[{id:"cq",name:"重庆",members:[{profileId:"profile-123",role:"owner"}],weknora:{tenantId:"42",apiKeyEnv:"WK_TEST"},openviking:{accountId:"cq"}}]}));
  process.env.WK_TEST="service-key";
  const config=resolveKnowledgeConfig({weknoraBaseUrl:"http://weknora.test",graphApiBaseUrl:"http://graph.test",workspaceRegistryPath:rp,workspaceStatePath:path.join(dir,"state.json"),uploadMaxBytes:1024*1024,requestTimeoutMs:5000});
  const workspaces=new WorkspaceRegistry(config.workspaceRegistryPath,config.workspaceStatePath);
  const identity={authenticatedUserProfile:{profileId:"profile-123",displayName:"Lee"}};
  return {config,workspaces,identity,principal:knowledgePrincipal(identity,workspaces)};
}

test("OpenClaw durable profile remains the only human identity",()=>{ const {identity}=setup(); assert.equal(requireDurableProfileId(identity),"profile-123"); assert.throws(()=>requireDurableProfileId({authenticatedUserId:"x"}),/durable OpenClaw/); });

test("v0.7 config removes static workspace, tenant and user identity",()=>{ const {config}=setup(); for(const key of ["weknoraApiKey","weknoraTenantId","workspaceId","externalUserId","weknoraBearerToken"]) assert.equal(key in config,false); assert.ok(config.workspaceRegistryPath); });

test("WeKnora headers are derived from resolved workspace plus OpenClaw profile",async()=>{ const {config,principal}=setup(); const client=new WeKnoraClient(config); await withFetch(async(url,init)=>{assert.equal(String(url),"http://weknora.test/api/v1/knowledge-bases"); assert.equal(init.headers["X-API-Key"],"service-key"); assert.equal(init.headers["X-Tenant-ID"],"42"); assert.equal(init.headers["X-External-User-ID"],"profile-123"); assert.equal(init.headers.Authorization,undefined); return jsonResponse({data:[{id:"kb1"}]});},async()=>{assert.equal((await client.listKnowledgeBases(principal))[0].id,"kb1");});});

test("knowledge lifecycle routes map to WeKnora contracts",async()=>{ const {config,principal}=setup(); const client=new WeKnoraClient(config); const seen=[]; await withFetch(async(url,init)=>{seen.push([init.method,new URL(String(url)).pathname]); return jsonResponse({success:true,data:{id:"x"}});},async()=>{await client.updateKnowledgeBase(principal,"kb1",{name:"A",description:"B"}); await client.deleteKnowledgeBase(principal,"kb1"); await client.reparseDocument(principal,"doc1"); await client.deleteDocument(principal,"doc1"); await client.createShare(principal,"kb1","org1","viewer");}); assert.deepEqual(seen,[["PUT","/api/v1/knowledge-bases/kb1"],["DELETE","/api/v1/knowledge-bases/kb1"],["POST","/api/v1/knowledge/doc1/reparse"],["DELETE","/api/v1/knowledge/doc1"],["POST","/api/v1/knowledge-bases/kb1/shares"]]);});

test("file upload becomes multipart and never exposes downstream credential to browser",async()=>{ const {config,principal}=setup(); const client=new WeKnoraClient(config); await withFetch(async(url,init)=>{assert.equal(new URL(String(url)).pathname,"/api/v1/knowledge-bases/kb1/knowledge/file"); assert.ok(init.body instanceof FormData); assert.equal(init.headers["Content-Type"],undefined); assert.equal(init.headers["X-API-Key"],"service-key"); return jsonResponse({success:true,data:{id:"doc1"}});},async()=>{const out=await client.uploadDocument(principal,"kb1",{filename:"a.txt",contentType:"text/plain",base64:Buffer.from("hello").toString("base64")}); assert.equal(out.id,"doc1");});});

test("ontology graph keeps stable GraphView endpoint",async()=>{ const {config,principal}=setup(); const client=new WeKnoraClient(config); await withFetch(async(url,init)=>{assert.equal(String(url),"http://graph.test/api/v1/knowledge-bases/kb1/graph?view=ontology&limit=240"); assert.equal(init.headers["X-External-User-ID"],"profile-123"); return jsonResponse({nodes:[],edges:[],meta:{view:"ontology"}});},async()=>{assert.equal((await client.graph(principal,"kb1","ontology")).meta.view,"ontology");});});

test("hierarchical wiki slugs preserve path separators",async()=>{ const {config,principal}=setup(); const client=new WeKnoraClient(config); await withFetch(async(url,init)=>{assert.equal(new URL(String(url)).pathname,"/api/v1/knowledgebase/kb1/wiki/pages/folder/page%20one"); return jsonResponse({id:"w1"});},async()=>{await client.updateWikiPage(principal,"kb1","folder/page one",{title:"Page"});});});
