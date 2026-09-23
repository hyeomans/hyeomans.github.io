const reloadButton = document.querySelector("#reload");

async function api(path) {
  const response = await fetch(path);
  const isJSON = response.headers.get("content-type")?.includes("application/json");
  const body = isJSON ? await response.json() : {};
  if (!response.ok) throw new Error(body.error || `Request failed: ${response.status}`);
  return body;
}

function definitionList(rows) {
  const list = document.createElement("dl");
  for (const [label, value] of rows) {
    const term = document.createElement("dt");
    const detail = document.createElement("dd");
    term.textContent = label;
    detail.textContent = value || "Not available";
    list.append(term, detail);
  }
  return list;
}

function renderApp(target, app) {
  const permissions = Object.entries(app.permissions || {})
    .map(([name, level]) => `${name}: ${level}`)
    .join(", ");
  target.replaceChildren(definitionList([
    ["Name", app.name],
    ["Slug", app.slug],
    ["Owner", app.owner?.login],
    ["Client ID", app.client_id],
    ["Permissions", permissions],
  ]));
}

function renderOrganization(target, org) {
  target.replaceChildren(definitionList([
    ["Account", org.login],
    ["Name", org.name],
    ["Description", org.description],
    ["Public repos", String(org.public_repos)],
  ]));
}

function renderRepositories(target, payload) {
  const list = document.createElement("ul");
  for (const repository of payload.repositories || []) {
    const item = document.createElement("li");
    const name = document.createElement("span");
    const visibility = document.createElement("small");
    name.textContent = repository.full_name;
    visibility.textContent = repository.private ? "private" : "public";
    item.append(name, visibility);
    list.append(item);
  }
  if (!list.children.length) {
    target.textContent = "No repositories are available to this installation.";
    return;
  }
  target.replaceChildren(list);
}

async function load(targetSelector, path, render) {
  const target = document.querySelector(targetSelector);
  delete target.dataset.kind;
  target.className = "state";
  target.textContent = "Loading…";
  try {
    render(target, await api(path));
    target.className = "";
  } catch (error) {
    target.dataset.kind = "error";
    target.textContent = error.message;
  }
}

async function loadDashboard() {
  reloadButton.disabled = true;
  await Promise.all([
    load("#app-output", "/api/app", renderApp),
    load("#org-output", "/api/organization", renderOrganization),
    load("#repos-output", "/api/installation/repositories", renderRepositories),
  ]);
  reloadButton.disabled = false;
}

reloadButton.addEventListener("click", loadDashboard);
loadDashboard();
