// Ejecutar: node --test frontend/
const test = require("node:test");
const assert = require("node:assert");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");

const code = fs.readFileSync(path.join(__dirname, "app.js"), "utf8");

function el(tag) {
  return {
    tag, children: [], dataset: {}, className: "", textContent: "", listeners: {},
    hidden: false, checked: false, value: "",
    append(...c) { this.children.push(...c); },
    appendChild(c) { this.children.push(c); },
    addEventListener(t, f) { this.listeners[t] = f; },
    set innerHTML(_) { this.children = []; },
    get innerHTML() { return ""; },
    reset() { this.resetCalled = true; },
    querySelectorAll(sel) {
      const out = [];
      const walk = (n) => {
        for (const c of n.children || []) {
          if (typeof c === "object") {
            if (c.tag === "input" && (!sel.includes(":checked") || c.checked)) out.push(c);
            walk(c);
          }
        }
      };
      walk(this);
      return out;
    },
  };
}

const IDS = [
  "lista-notas", "nota-form", "lista-categorias", "categoria-form",
  "categorias-select", "btn-agregar-nota", "btn-agregar-categoria",
];

function setup(responses) {
  const calls = [];
  const els = Object.fromEntries(IDS.map((id) => [id, el(id)]));
  const nota = els["nota-form"];
  nota.titulo = { value: "T" };
  nota.contenido = { value: "C" };
  els["categoria-form"].nombre = { value: "Trabajo" };
  const doc = {
    getElementById: (id) => els[id],
    createElement: el,
    addEventListener() {},
  };
  const fetch = async (url, opts = {}) => {
    calls.push({ url, method: opts.method || "GET", body: opts.body });
    const r = responses.shift() || { ok: true, json: [] };
    return { ok: r.ok, status: r.ok ? 200 : 500, json: async () => r.json };
  };
  const alerts = [];
  const ctx = vm.createContext({ document: doc, fetch, alert: (m) => alerts.push(m) });
  vm.runInContext(code + "\n;this.cargarNotas = cargarNotas; this.cargarCategorias = cargarCategorias;", ctx);
  return { ctx, calls, els, lista: els["lista-notas"], form: nota, alerts };
}

const tick = () => new Promise((r) => setImmediate(r));

test("cargarNotas hace GET y renderiza un li por nota con botón eliminar", async () => {
  const { ctx, calls, lista } = setup([
    { ok: true, json: [
      { id_nota: 1, titulo: "A", contenido: { String: "x", Valid: true }, categorias: [] },
      { id_nota: 2, titulo: "B", contenido: "y", categorias: [] },
    ] },
  ]);
  await ctx.cargarNotas();
  assert.deepStrictEqual(calls[0], { url: "/api/notas", method: "GET", body: undefined });
  assert.strictEqual(lista.children.length, 2);
  const btn = lista.children[0].children[1].children[1];
  assert.strictEqual(btn.dataset.id, 1);
  assert.strictEqual(btn.className, "eliminar");
  assert.strictEqual(lista.children[0].children[0].children[1].textContent, "x");
  assert.strictEqual(lista.children[1].children[0].children[1].textContent, "y");
});

test("cada nota muestra sus categorías como chips", async () => {
  const { ctx, lista } = setup([
    { ok: true, json: [{ id_nota: 1, titulo: "A", contenido: "", categorias: [
      { id_categoria: 1, nombre: "Casa" }, { id_categoria: 2, nombre: "Trabajo" },
    ] }] },
  ]);
  await ctx.cargarNotas();
  const chips = lista.children[0].children[0].children.filter((c) => c.className === "chip");
  assert.deepStrictEqual(chips.map((c) => c.textContent), ["Casa", "Trabajo"]);
});

test("cargarCategorias hace GET, muestra chips y checkboxes", async () => {
  const { ctx, calls, els } = setup([
    { ok: true, json: [{ id_categoria: 1, nombre: "Casa" }, { id_categoria: 2, nombre: "Trabajo" }] },
  ]);
  await ctx.cargarCategorias();
  assert.strictEqual(calls[0].url, "/api/categorias");
  const chips = els["lista-categorias"].children;
  assert.deepStrictEqual(chips.map((c) => c.children[0].textContent), ["Casa", "Trabajo"]);
  assert.ok(chips.every((c) => c.className === "chip"));
  const inputs = els["categorias-select"].querySelectorAll("input");
  assert.deepStrictEqual(inputs.map((i) => i.value), [1, 2]);
});

test("cargarCategorias tolera respuesta null", async () => {
  const { ctx, els } = setup([{ ok: true, json: null }]);
  await ctx.cargarCategorias();
  assert.strictEqual(els["lista-categorias"].children.length, 0);
});

test("botón agregar categoría despliega y oculta el formulario", () => {
  const { els } = setup([]);
  els["categoria-form"].hidden = true;
  els["btn-agregar-categoria"].listeners.click();
  assert.strictEqual(els["categoria-form"].hidden, false);
  els["btn-agregar-categoria"].listeners.click();
  assert.strictEqual(els["categoria-form"].hidden, true);
});

test("botón agregar nota despliega y oculta el formulario", () => {
  const { els } = setup([]);
  els["nota-form"].hidden = true;
  els["btn-agregar-nota"].listeners.click();
  assert.strictEqual(els["nota-form"].hidden, false);
  els["btn-agregar-nota"].listeners.click();
  assert.strictEqual(els["nota-form"].hidden, true);
});

test("submit de categoría hace POST con nombre y recarga categorías", async () => {
  const { calls, els } = setup([{ ok: true, json: {} }, { ok: true, json: [] }]);
  let prevented = false;
  await els["categoria-form"].listeners.submit({ preventDefault: () => (prevented = true) });
  await tick();
  assert.ok(prevented);
  assert.strictEqual(calls[0].url, "/api/categorias");
  assert.strictEqual(calls[0].method, "POST");
  assert.deepStrictEqual(JSON.parse(calls[0].body), { nombre: "Trabajo" });
  assert.ok(els["categoria-form"].resetCalled);
  assert.strictEqual(calls[1].url, "/api/categorias");
  assert.strictEqual(calls[1].method, "GET");
});

test("submit de nota previene default, hace POST con categorías elegidas y refresca", async () => {
  const { calls, form, els, ctx } = setup([
    { ok: true, json: [{ id_categoria: 1, nombre: "A" }, { id_categoria: 2, nombre: "B" }] },
    { ok: true, json: {} },
    { ok: true, json: [] },
  ]);
  await ctx.cargarCategorias();
  els["categorias-select"].querySelectorAll("input")[1].checked = true;
  let prevented = false;
  await form.listeners.submit({ preventDefault: () => (prevented = true) });
  await tick();
  assert.ok(prevented);
  assert.strictEqual(calls[1].method, "POST");
  assert.deepStrictEqual(JSON.parse(calls[1].body), {
    titulo: "T", contenido: "C", is_archived: false, categorias: [2],
  });
  assert.ok(form.resetCalled);
  assert.strictEqual(calls[2].method, "GET");
});

test("click en eliminar (delegación) hace DELETE con el id y refresca", async () => {
  const { calls, lista } = setup([{ ok: true, json: [] }, { ok: true, json: [] }]);
  const btn = { dataset: { id: "7" } };
  await lista.listeners.click({ target: { closest: (q) => (q === ".eliminar" ? btn : null) } });
  await tick();
  assert.strictEqual(calls[0].url, "/api/notas/7");
  assert.strictEqual(calls[0].method, "DELETE");
  assert.strictEqual(calls[1].method, "GET");
});

test("click fuera de un botón eliminar no hace nada", async () => {
  const { calls, lista } = setup([]);
  await lista.listeners.click({ target: { closest: () => null } });
  assert.strictEqual(calls.length, 0);
});

test("error en DELETE muestra alert y no refresca", async () => {
  const { calls, lista, alerts } = setup([{ ok: false }]);
  await lista.listeners.click({ target: { closest: (q) => (q === ".eliminar" ? { dataset: { id: "1" } } : null) } });
  assert.strictEqual(calls.length, 1);
  assert.strictEqual(alerts.length, 1);
});

const CATS = [{ id_categoria: 1, nombre: "A" }, { id_categoria: 2, nombre: "B" }];
const NOTA = { id_nota: 5, titulo: "Vieja", contenido: { String: "txt", Valid: true }, categorias: [{ id_categoria: 2, nombre: "B" }] };
const sel = (s) => (e) => (e === s ? { dataset: { id: "5" } } : null);

test("cada nota tiene botón editar", async () => {
  const { ctx, lista } = setup([{ ok: true, json: [NOTA] }]);
  await ctx.cargarNotas();
  const ed = lista.children[0].children[1].children[0];
  assert.strictEqual(ed.className, "editar");
  assert.strictEqual(ed.dataset.id, 5);
});

test("click en editar carga la nota en el formulario", async () => {
  const { ctx, lista, form, els } = setup([
    { ok: true, json: CATS }, { ok: true, json: [NOTA] },
  ]);
  await ctx.cargarCategorias();
  await ctx.cargarNotas();
  form.hidden = true;
  await lista.listeners.click({ target: { closest: sel(".editar") } });
  assert.strictEqual(form.hidden, false);
  assert.strictEqual(form.titulo.value, "Vieja");
  assert.strictEqual(form.contenido.value, "txt");
  const inputs = els["categorias-select"].querySelectorAll("input");
  assert.deepStrictEqual(inputs.map((i) => i.checked), [false, true]);
});

test("submit en modo edición hace PUT y luego vuelve a POST", async () => {
  const { ctx, lista, form, calls } = setup([
    { ok: true, json: CATS }, { ok: true, json: [NOTA] },
    { ok: true, json: {} }, { ok: true, json: [] },
    { ok: true, json: {} }, { ok: true, json: [] },
  ]);
  await ctx.cargarCategorias();
  await ctx.cargarNotas();
  await lista.listeners.click({ target: { closest: sel(".editar") } });
  await form.listeners.submit({ preventDefault() {} });
  await tick();
  assert.strictEqual(calls[2].url, "/api/notas/5");
  assert.strictEqual(calls[2].method, "PUT");
  assert.deepStrictEqual(JSON.parse(calls[2].body), {
    titulo: "Vieja", contenido: "txt", is_archived: false, categorias: [2],
  });
  assert.strictEqual(calls[3].method, "GET");
  await form.listeners.submit({ preventDefault() {} });
  await tick();
  assert.strictEqual(calls[4].method, "POST");
});

test("chips de categoría tienen botones editar y eliminar", async () => {
  const { ctx, els } = setup([{ ok: true, json: CATS }]);
  await ctx.cargarCategorias();
  const chip = els["lista-categorias"].children[0];
  assert.strictEqual(chip.children[1].className, "editar-cat");
  assert.strictEqual(chip.children[2].className, "eliminar-cat");
  assert.strictEqual(chip.children[2].dataset.id, 1);
});

test("editar categoría rellena el formulario y envía PUT", async () => {
  const { ctx, els, calls } = setup([
    { ok: true, json: CATS }, { ok: true, json: {} }, { ok: true, json: [] },
  ]);
  await ctx.cargarCategorias();
  const cf = els["categoria-form"];
  cf.hidden = true;
  await els["lista-categorias"].listeners.click({
    target: { closest: (s) => (s === ".editar-cat" ? { dataset: { id: "2" } } : null) },
  });
  assert.strictEqual(cf.hidden, false);
  assert.strictEqual(cf.nombre.value, "B");
  await cf.listeners.submit({ preventDefault() {} });
  await tick();
  assert.strictEqual(calls[1].url, "/api/categorias/2");
  assert.strictEqual(calls[1].method, "PUT");
  assert.deepStrictEqual(JSON.parse(calls[1].body), { nombre: "B" });
  assert.strictEqual(calls[2].method, "GET");
});

test("eliminar categoría hace DELETE y refresca categorías y notas", async () => {
  const { els, calls } = setup([
    { ok: true, json: {} }, { ok: true, json: [] }, { ok: true, json: [] },
  ]);
  await els["lista-categorias"].listeners.click({
    target: { closest: (s) => (s === ".eliminar-cat" ? { dataset: { id: "3" } } : null) },
  });
  await tick();
  assert.strictEqual(calls[0].url, "/api/categorias/3");
  assert.strictEqual(calls[0].method, "DELETE");
  assert.deepStrictEqual(calls.slice(1).map((c) => c.url).sort(), ["/api/categorias", "/api/notas"]);
});

test("error al eliminar categoría muestra alert", async () => {
  const { els, calls, alerts } = setup([{ ok: false }]);
  await els["lista-categorias"].listeners.click({
    target: { closest: (s) => (s === ".eliminar-cat" ? { dataset: { id: "3" } } : null) },
  });
  assert.strictEqual(calls.length, 1);
  assert.strictEqual(alerts.length, 1);
});

test("editar una categoría también refresca las notas", async () => {
  const { ctx, els, calls } = setup([
    { ok: true, json: CATS }, { ok: true, json: {} }, { ok: true, json: [] }, { ok: true, json: [] },
  ]);
  await ctx.cargarCategorias();
  await els["lista-categorias"].listeners.click({
    target: { closest: (s) => (s === ".editar-cat" ? { dataset: { id: "2" } } : null) },
  });
  await els["categoria-form"].listeners.submit({ preventDefault() {} });
  await tick();
  assert.deepStrictEqual(calls.slice(2).map((c) => c.url).sort(), ["/api/categorias", "/api/notas"]);
});
