import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { OpenVikingClient, resolveOpenVikingConfig } from "../lib/client.js";
import { openVikingPrincipal, sessionPrincipal } from "../lib/principal.js";
import { WorkspaceRegistry } from "../../workspace-core/index.js";
import { latestTurn, openVikingSessionId, renderMemoryContext } from "../lib/memory-runtime.js";

function jsonResponse(body,status=200){return new Response(JSON.stringify(body),{status,headers:{"Content-Type":"application/json"}})}
async function withFetch(handler,fn){const old=globalThis.fetch;globalThis.fetch=handler;try{await fn();}finally{globalThis.fetch=old;}}
function setup(){const dir=fs.mkdtempSync(path.join(os.tmpdir(),"leeclaw-ov-")); const rp=path.join(dir,"workspaces.json"); fs.writeFileSync(rp,JSON.stringify({workspaces:[{id:"cq",name:"重庆",members:[{profileId:"p1",role:"owner"}],weknora:{tenantId:"42",apiKeyEnv:"WK_OV"},openviking:{accountId:"ov-cq"}},{id:"hq",name:"总部",members:[{profileId:"p1",role:"editor"}],weknora:{tenantId:"1",apiKeyEnv:"WK_HQ_OV"},openviking:{accountId:"ov-hq"}}]})); delete process.env.WK_OV; delete process.env.WK_HQ_OV; const config=resolveOpenVikingConfig({baseUrl:"http://ov.test",apiKey:"root",workspaceRegistryPath:rp,workspaceStatePath:path.join(dir,"state.json"),requestTimeoutMs:5000}); const registry=new WorkspaceRegistry(config.workspaceRegistryPath,config.workspaceStatePath); return {config,registry};}

test("OpenViking principal resolves account from validated active workspace",()=>{const {config,registry}=setup(); const client={authenticatedUserProfile:{profileId:"p1"}}; assert.equal(openVikingPrincipal(client,registry).accountId,"ov-cq"); registry.switch("p1","hq"); assert.equal(openVikingPrincipal(client,registry).accountId,"ov-hq"); assert.equal("workspaceId" in config,false);});

test("OpenViking trusted headers never accept browser user/account",async()=>{const {config,registry}=setup(); const client=new OpenVikingClient(config); const principal=openVikingPrincipal({authenticatedUserProfile:{profileId:"p1"}},registry); await withFetch(async(url,init)=>{assert.equal(init.headers["X-OpenViking-Account"],"ov-cq");assert.equal(init.headers["X-OpenViking-User"],"p1");return jsonResponse({result:{skills:[]}});},()=>client.listSkills(principal));});

test("session principal follows workspace switch for the session owner profile",()=>{const {registry}=setup(); const api={runtime:{agent:{session:{getSessionEntry(){return{createdActor:{type:"human",source:"profile",id:"p1"}}}}}}}; assert.equal(sessionPrincipal(api,registry,"s1").accountId,"ov-cq"); registry.switch("p1","hq"); assert.equal(sessionPrincipal(api,registry,"s1").accountId,"ov-hq");});

test("memory helper remains deterministic and marks memory non-authoritative",()=>{const turn=latestTurn([{role:"user",content:"q"},{role:"assistant",content:"a"}]);assert.deepEqual(turn,{user:"q",assistant:"a"}); const p={workspaceId:"cq",userId:"p1"}; assert.equal(openVikingSessionId(p,"s1"),openVikingSessionId(p,"s1")); const ctx=renderMemoryContext({memories:[{abstract:"remember me",uri:"viking://x"}]}); assert.match(ctx,/not authoritative enterprise facts/);});
