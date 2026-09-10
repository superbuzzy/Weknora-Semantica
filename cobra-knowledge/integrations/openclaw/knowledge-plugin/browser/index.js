import { defineControlUiPlugin } from "openclaw/plugin-sdk/control-ui";
import "./knowledge.css";

function el(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
}

function text(value, fallback = "—") {
  if (value === undefined || value === null || value === "") return fallback;
  return String(value);
}

function pickName(item) {
  return text(item?.name ?? item?.title ?? item?.filename ?? item?.file_name ?? item?.id, "未命名");
}

function pickStatus(item) {
  return text(item?.parse_status ?? item?.status ?? item?.state ?? item?.role, "");
}

function renderRows(container, rows, kind = "generic") {
  container.replaceChildren();
  if (!rows?.length) {
    container.append(el("div", "lk-empty", "暂无数据"));
    return;
  }
  for (const item of rows) {
    const row = el("div", "lk-row");
    const main = el("div", "lk-row-main");
    main.append(el("div", "lk-row-title", pickName(item)));
    const secondary = kind === "wiki"
      ? item?.summary
      : item?.description ?? item?.email ?? item?.uri ?? item?.source ?? item?.id;
    if (secondary) main.append(el("div", "lk-row-sub", String(secondary)));
    row.append(main);
    const status = pickStatus(item);
    if (status) row.append(el("span", "lk-badge", status));
    container.append(row);
  }
}

function renderGraph(container, graph) {
  container.replaceChildren();
  const nodes = Array.isArray(graph?.nodes) ? graph.nodes : [];
  const edges = Array.isArray(graph?.edges) ? graph.edges : [];
  if (!nodes.length) {
    container.append(el("div", "lk-empty", "暂无图谱节点"));
    return;
  }
  const width = Math.max(760, container.clientWidth || 760);
  const height = 480;
  const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  svg.setAttribute("viewBox", `0 0 ${width} ${height}`);
  svg.classList.add("lk-graph-svg");
  const radius = Math.min(width, height) * 0.34;
  const cx = width / 2;
  const cy = height / 2;
  const positions = new Map();
  nodes.forEach((node, index) => {
    const angle = (Math.PI * 2 * index) / nodes.length - Math.PI / 2;
    const ring = nodes.length <= 3 ? radius * 0.55 : radius;
    positions.set(node.id, { x: cx + Math.cos(angle) * ring, y: cy + Math.sin(angle) * ring });
  });
  for (const edge of edges) {
    const a = positions.get(edge.source);
    const b = positions.get(edge.target);
    if (!a || !b) continue;
    const line = document.createElementNS(svg.namespaceURI, "line");
    line.setAttribute("x1", a.x); line.setAttribute("y1", a.y);
    line.setAttribute("x2", b.x); line.setAttribute("y2", b.y);
    line.setAttribute("class", "lk-edge");
    svg.append(line);
  }
  for (const node of nodes) {
    const p = positions.get(node.id);
    const group = document.createElementNS(svg.namespaceURI, "g");
    group.setAttribute("transform", `translate(${p.x} ${p.y})`);
    const circle = document.createElementNS(svg.namespaceURI, "circle");
    circle.setAttribute("r", node.kind === "property" ? "18" : "24");
    circle.setAttribute("class", `lk-node lk-node-${node.kind || "entity"}`);
    const label = document.createElementNS(svg.namespaceURI, "text");
    label.setAttribute("y", node.kind === "property" ? "34" : "41");
    label.setAttribute("class", "lk-node-label");
    label.textContent = text(node.label ?? node.id);
    group.append(circle, label);
    svg.append(group);
  }
  container.append(svg);
  const meta = graph?.meta;
  if (meta) {
    container.append(el("div", "lk-graph-meta", `${text(meta.view, "graph")} · ${text(meta.returned_nodes ?? nodes.length)} 节点 · ${text(meta.returned_edges ?? edges.length)} 边${meta.ontology_version ? ` · ontology ${meta.ontology_version}` : ""}`));
  }
}

function createKnowledgePage(host) {
  return (container, context) => {
    let selectedKb = null;
    let activeTab = "documents";
    let disposed = false;

    const root = el("section", "lk-page");
    const header = el("header", "lk-header");
    const headingWrap = el("div");
    headingWrap.append(el("h1", "lk-title", "Knowledge"), el("p", "lk-subtitle", "WeKnora-backed enterprise knowledge, managed natively inside OpenClaw."));
    const actions = el("div", "lk-actions");
    const refresh = el("button", "lk-button lk-button-secondary", "刷新");
    const create = el("button", "lk-button", "新建知识库");
    actions.append(refresh, create);
    header.append(headingWrap, actions);

    const layout = el("div", "lk-layout");
    const sidebar = el("aside", "lk-sidebar");
    const sideHead = el("div", "lk-side-head", "知识库");
    const list = el("div", "lk-kb-list");
    sidebar.append(sideHead, list);

    const main = el("main", "lk-main");
    const empty = el("div", "lk-empty-panel", "选择一个知识库查看文档、Wiki、图谱和权限。");
    const detail = el("div", "lk-detail");
    detail.hidden = true;
    const detailHead = el("div", "lk-detail-head");
    const detailTitle = el("h2", "lk-detail-title");
    const detailDesc = el("p", "lk-detail-desc");
    detailHead.append(detailTitle, detailDesc);
    const tabs = el("div", "lk-tabs");
    const content = el("div", "lk-content");
    detail.append(detailHead, tabs, content);
    main.append(empty, detail);
    layout.append(sidebar, main);
    root.append(header, layout);
    container.append(root);

    const call = (method, params = {}) => host.request(method, params);

    const setBusy = (busy) => {
      refresh.disabled = busy;
      create.disabled = busy;
    };

    function drawKbList(items) {
      list.replaceChildren();
      if (!items.length) {
        list.append(el("div", "lk-empty", "暂无知识库"));
        return;
      }
      for (const kb of items) {
        const button = el("button", `lk-kb${selectedKb?.id === kb.id ? " is-active" : ""}`);
        button.append(el("strong", "lk-kb-name", pickName(kb)));
        const meta = [kb.type, kb.vector_store_source, kb.knowledge_count ?? kb.total].filter(Boolean).join(" · ");
        if (meta) button.append(el("span", "lk-kb-meta", meta));
        button.onclick = () => selectKb(kb);
        list.append(button);
      }
    }

    async function loadKbList() {
      setBusy(true);
      try {
        const items = await call("leeclaw.knowledge.list", { creator: "all" });
        if (!disposed) drawKbList(Array.isArray(items) ? items : []);
      } catch (error) {
        if (!disposed) {
          list.replaceChildren(el("div", "lk-error", String(error)));
        }
      } finally {
        if (!disposed) setBusy(false);
      }
    }

    function drawTabs() {
      const tabDefs = [
        ["documents", "文档"],
        ["wiki", "Wiki"],
        ["entity", "实体图"],
        ["ontology", "本体图"],
        ["access", "共享与权限"],
        ["activity", "审计"],
      ];
      tabs.replaceChildren();
      for (const [id, label] of tabDefs) {
        const button = el("button", `lk-tab${activeTab === id ? " is-active" : ""}`, label);
        button.onclick = () => { activeTab = id; drawTabs(); loadActiveTab(); };
        tabs.append(button);
      }
    }

    async function selectKb(kb) {
      selectedKb = kb;
      empty.hidden = true;
      detail.hidden = false;
      detailTitle.textContent = pickName(kb);
      detailDesc.textContent = text(kb.description, `ID: ${kb.id}`);
      drawTabs();
      await loadActiveTab();
      loadKbList();
    }

    async function loadActiveTab() {
      if (!selectedKb) return;
      content.replaceChildren(el("div", "lk-loading", "加载中…"));
      try {
        if (activeTab === "documents") {
          const result = await call("leeclaw.knowledge.documents", { kbId: selectedKb.id, page: 1, page_size: 100 });
          renderRows(content, result?.items ?? [], "document");
        } else if (activeTab === "wiki") {
          const result = await call("leeclaw.knowledge.wiki", { kbId: selectedKb.id, page: 1, page_size: 100 });
          renderRows(content, result?.items ?? [], "wiki");
        } else if (activeTab === "entity" || activeTab === "ontology") {
          const result = await call("leeclaw.knowledge.graph", { kbId: selectedKb.id, view: activeTab });
          renderGraph(content, result);
        } else if (activeTab === "access") {
          const access = el("div", "lk-access-grid");
          const membersCard = el("section", "lk-card");
          membersCard.append(el("h3", "lk-card-title", "工作空间成员"));
          const membersBody = el("div");
          membersCard.append(membersBody);
          const sharesCard = el("section", "lk-card");
          sharesCard.append(el("h3", "lk-card-title", "知识库共享"));
          const sharesBody = el("div");
          sharesCard.append(sharesBody);
          access.append(membersCard, sharesCard);
          content.replaceChildren(access);
          const [members, shares] = await Promise.all([
            call("leeclaw.knowledge.members", {}),
            call("leeclaw.knowledge.shares", { kbId: selectedKb.id }),
          ]);
          renderRows(membersBody, members?.items ?? []);
          renderRows(sharesBody, shares?.items ?? []);
        } else if (activeTab === "activity") {
          const result = await call("leeclaw.knowledge.activity", { kbId: selectedKb.id });
          renderRows(content, result?.items ?? []);
        }
      } catch (error) {
        if (!disposed) content.replaceChildren(el("div", "lk-error", String(error)));
      }
    }

    refresh.onclick = () => selectedKb ? loadActiveTab() : loadKbList();
    create.onclick = async () => {
      const name = window.prompt("知识库名称");
      if (!name?.trim()) return;
      const description = window.prompt("描述（可选）") ?? "";
      setBusy(true);
      try {
        const kb = await call("leeclaw.knowledge.create", { name: name.trim(), description });
        await loadKbList();
        if (kb?.id) await selectKb(kb);
      } catch (error) {
        window.alert(String(error));
      } finally {
        setBusy(false);
      }
    };

    loadKbList();
    return {
      dispose() {
        disposed = true;
        root.remove();
      },
    };
  };
}

export default defineControlUiPlugin({
  id: "leeclaw-knowledge",
  activate(host) {
    const page = host.ui.registerPage({ id: "knowledge", label: "Knowledge", mount: createKnowledgePage(host) });
    const nav = host.ui.registerNavigation({ id: "knowledge", label: "Knowledge", page: { id: "knowledge" }, icon: "bookOpen", order: 30 });
    return () => { nav(); page(); };
  },
});
