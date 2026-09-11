import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { OpenVikingClient, resolveOpenVikingConfig } from "../lib/client.js";
import { openVikingPrincipal, sessionPrincipal } from "../lib/principal.js";
import { WorkspaceRegistry } from "../../workspace-core/index.js";
import { latestTurn, openVikingSessionId, renderMemoryContext } from "../lib/memory-runtime.js";
import { parseResolvedSkill, renderSkillContext, resolveSkillForTurn } from "../lib/skill-runtime.js";
import { registerLeeClawAgentRuntime } from "../lib/agent-runtime.js";

function jsonResponse(body,status=200){return new Response(JSON.stringify(body),{status,headers:{"Content-Type":"application/json"}})}
async function withFetch(handler,fn){const old=globalThis.fetch;globalThis.fetch=handler;try{return await fn();}finally{globalThis.fetch=old;}}
function setup(){
  const dir=fs.mkdtempSync(path.join(os.tmpdir(),"leeclaw-ov-")); const rp=path.join(dir,"workspaces.json");
  fs.writeFileSync(rp,JSON.stringify({workspaces:[
    {id:"cq",name:"重庆",members:[{profileId:"p1",role:"owner"}],weknora:{tenantId:"42",apiKeyEnv:"WK_OV"},openviking:{accountId:"ov-cq"}},
    {id:"hq",name:"总部",members:[{profileId:"p1",role:"editor"}],weknora:{tenantId:"1",apiKeyEnv:"WK_HQ_OV"},openviking:{accountId:"ov-hq"}}
  ]}));
  const config=resolveOpenVikingConfig({baseUrl:"http://ov.test",apiKey:"root",workspaceRegistryPath:rp,workspaceStatePath:path.join(dir,"state.json"),requestTimeoutMs:5000,skillScoreThreshold:0.7,skillResolveLimit:3,skillMaxContentChars:8000});
  const registry=new WorkspaceRegistry(config.workspaceRegistryPath,config.workspaceStatePath);
  const api={runtime:{agent:{session:{getSessionEntry(){return{createdActor:{type:"human",source:"profile",id:"p1"}}}}}}};
  return {config,registry,api};
}

test("OpenViking principal resolves account from validated active workspace",()=>{const {config,registry}=setup(); const client={authenticatedUserProfile:{profileId:"p1"}}; assert.equal(openVikingPrincipal(client,registry).accountId,"ov-cq"); registry.switch("p1","hq"); assert.equal(openVikingPrincipal(client,registry).accountId,"ov-hq"); assert.equal("workspaceId" in config,false);});

test("OpenViking trusted headers never accept browser user/account",async()=>{const {config,registry}=setup(); const client=new OpenVikingClient(config); const principal=openVikingPrincipal({authenticatedUserProfile:{profileId:"p1"}},registry); await withFetch(async(_url,init)=>{assert.equal(init.headers["X-OpenViking-Account"],"ov-cq");assert.equal(init.headers["X-OpenViking-User"],"p1");return jsonResponse({result:{skills:[]}});},()=>client.listSkills(principal));});

test("session principal remains pinned while a new session adopts the switched workspace",()=>{const {registry,api}=setup(); assert.equal(sessionPrincipal(api,registry,"s1","sid-1").accountId,"ov-cq"); registry.switch("p1","hq"); assert.equal(sessionPrincipal(api,registry,"s1","sid-1").accountId,"ov-cq"); assert.equal(sessionPrincipal(api,registry,"s1","sid-2").accountId,"ov-hq");});

test("memory helper remains deterministic and marks memory non-authoritative",()=>{const turn=latestTurn([{role:"user",content:"q"},{role:"assistant",content:"a"}]);assert.deepEqual(turn,{user:"q",assistant:"a"}); const p={workspaceId:"cq",userId:"p1"}; assert.equal(openVikingSessionId(p,"s1"),openVikingSessionId(p,"s1")); const ctx=renderMemoryContext({memories:[{abstract:"remember me",uri:"viking://x"}]}); assert.match(ctx,/not authoritative enterprise facts/);});

test("Skill parser maps only safely translatable allowed tools through host authority",()=>{
  const authority={assertActive(){},allows(name){return ["leeclaw_context_retrieve","read","exec","web_search"].includes(name);}};
  const skill=parseResolvedSkill({name:"procurement",score:0.91,root_uri:"viking://agent/skills/procurement"},{skill_md:"---\nname: procurement\ndescription: test\nallowed-tools: context.retrieve Read Bash WebSearch Bash(git:*) unknown\n---\n\nDo the work."},authority);
  assert.equal(skill.allowedToolsDeclared,true);
  assert.deepEqual(skill.toolsAllow,["leeclaw_context_retrieve","read","exec","web_search"]);
  assert.equal(skill.content,"Do the work.");
});

test("explicit empty allowed-tools is a real deny-all optional tool policy",()=>{
  const skill=parseResolvedSkill({name:"review",score:0.8},{skill_md:"---\nname: review\ndescription: review\nallowed-tools:\n---\nOnly reason."},{assertActive(){},allows(){return true;}});
  assert.equal(skill.allowedToolsDeclared,true); assert.deepEqual(skill.toolsAllow,[]);
});


test("Skill parser treats underscore YAML allowed_tools and block lists as real host restrictions",()=>{
  const authority={assertActive(){},allows(name){return ["read","exec"].includes(name);}};
  const skill=parseResolvedSkill({name:"yaml",score:0.9},{content:"---\nname: yaml\ndescription: yaml\nallowed_tools:\n  - Read\n  - Bash\n---\nDo it."},authority);
  assert.equal(skill.allowedToolsDeclared,true);
  assert.deepEqual(skill.toolsAllow,["read","exec"]);
  const deny=parseResolvedSkill({name:"deny",score:0.9},{content:"---\nname: deny\ndescription: deny\nallowed_tools: []\n---\nReason only."},authority);
  assert.equal(deny.allowedToolsDeclared,true);
  assert.deepEqual(deny.toolsAllow,[]);
});

test("Skill resolver applies score threshold then loads level-2 skill content",async()=>{
  const {config,registry}=setup(); const principal=openVikingPrincipal({authenticatedUserProfile:{profileId:"p1"}},registry); const client=new OpenVikingClient(config); const calls=[];
  await withFetch(async(url,init)=>{
    calls.push([String(url),init.method,init.body]);
    if(String(url).includes("/skills/find")) return jsonResponse({result:{skills:[{name:"weak",score:0.2,root_uri:"viking://agent/skills/weak"},{name:"planning",score:0.92,root_uri:"viking://agent/skills/planning",match_reason:"semantic"}]}});
    return jsonResponse({result:{skill_md:"---\nname: planning\ndescription: plan\nallowed-tools: context.retrieve context.get_evidence\n---\nUse evidence."}});
  },async()=>{
    const skill=await resolveSkillForTurn(client,principal,"规划这个项目",config,{assertActive(){},allows(){return true;}});
    assert.equal(skill.name,"planning"); assert.equal(skill.score,0.92); assert.deepEqual(skill.toolsAllow,["leeclaw_context_retrieve","leeclaw_context_get_evidence"]);
  });
  const body=JSON.parse(calls[0][2]); assert.equal(body.score_threshold,0.7); assert.equal(body.limit,3); assert.match(calls[1][0],/target_uri=viking%3A%2F%2Fagent%2Fskills/);
});


test("Skill resolver prefers a personal Skill over a higher-scoring shared Skill",async()=>{
  const {config,registry}=setup(); const principal=openVikingPrincipal({authenticatedUserProfile:{profileId:"p1"}},registry);
  const loaded=[];
  const client={
    async findSkills(){return {skills:[
      {name:"shared-planning",score:0.96,root_uri:"viking://agent/skills/shared-planning"},
      {name:"my-planning",score:0.74,root_uri:"viking://user/p1/skills/my-planning"}
    ]};},
    async getSkill(_p,name,target){loaded.push([name,target]); return {skill_md:`---\nname: ${name}\ndescription: personal\n---\nUse personal SOP.`};}
  };
  const skill=await resolveSkillForTurn(client,principal,"规划",config,{assertActive(){},allows(){return true;}});
  assert.equal(skill.name,"my-planning");
  assert.deepEqual(loaded,[["my-planning","viking://user/p1/skills"]]);
});

test("Skill resolver returns null when no candidate reaches threshold",async()=>{
  const {config,registry}=setup(); const principal=openVikingPrincipal({authenticatedUserProfile:{profileId:"p1"}},registry);
  const client={async findSkills(){return {skills:[{name:"weak",score:0.2}]};},async getSkill(){throw new Error("must not load");}};
  assert.equal(await resolveSkillForTurn(client,principal,"q",config,{allows(){return true;}}),null);
});

test("Skill context is bounded and states the authority relationship",()=>{
  const text=renderSkillContext({name:"x",description:"d",sourceUri:"viking://x",score:0.9,content:"A".repeat(50000)},3000);
  assert.ok(text.length<=3000); assert.match(text,/host security, authorization, or tool policy/); assert.match(text,/Enterprise facts/);
});

test("one before_prompt_build hook composes Memory + Skill and narrows host tools",async()=>{
  const {config,registry}=setup(); const hooks=new Map(); let beforeOpts;
  const api={
    logger:{warn(){}}, runtime:{agent:{session:{getSessionEntry(){return{createdActor:{type:"human",source:"profile",id:"p1"}}}}}},
    on(name,handler,opts){hooks.set(name,handler); if(name==="before_prompt_build") beforeOpts=opts;}
  };
  const client={
    async searchMemory(){return {memories:[{abstract:"历史偏好"}]};},
    async findSkills(){return {skills:[{name:"planning",score:0.95,root_uri:"viking://agent/skills/planning"}]};},
    async getSkill(){return {skill_md:"---\nname: planning\ndescription: plan\nallowed-tools: context.retrieve\n---\nFollow planning SOP."};}
  };
  registerLeeClawAgentRuntime(api,config,client,registry);
  assert.equal(beforeOpts.requiresToolAuthority,true);
  const out=await hooks.get("before_prompt_build")({prompt:"做规划",messages:[]},{sessionKey:"s1",sessionId:"sid-runtime",toolAuthority:{assertActive(){},allows(name){return name==="leeclaw_context_retrieve";}}});
  assert.match(out.prependContext,/历史偏好/); assert.match(out.prependContext,/Follow planning SOP/); assert.deepEqual(out.toolsAllow,["leeclaw_context_retrieve"]);
});
