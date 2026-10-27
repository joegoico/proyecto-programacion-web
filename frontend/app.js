const API = "/api/notas";
const API_CAT = "/api/categorias";
const $ = (id) => document.getElementById(id);
const lista = $("lista-notas");
const form = $("nota-form");
const listaCategorias = $("lista-categorias");
const categoriaForm = $("categoria-form");
const categoriasSelect = $("categorias-select");
let notasCache = [];
let categoriasCache = [];
let editandoNota = null;
let editandoCategoria = null;

function boton(clase, id, texto) {
  const b = document.createElement("button");
  b.className = clase;
  b.dataset.id = id;
  b.textContent = texto;
  return b;
}

function acciones(claseEditar, claseEliminar, id) {
  const d = document.createElement("div");
  d.className = "acciones";
  d.append(boton(claseEditar, id, "Editar"), boton(claseEliminar, id, "Eliminar"));
  return d;
}

function chip(nombre) {
  const s = document.createElement("span");
  s.className = "chip";
  s.textContent = nombre;
  return s;
}

async function cargarNotas() {
  try {
    const res = await fetch(API);
    if (!res.ok) throw new Error(res.status);
    const notas = (await res.json()) || [];
    notasCache = notas;
    lista.innerHTML = "";
    for (const n of notas) {
      const contenido =
        n.contenido && typeof n.contenido === "object"
          ? n.contenido.String
          : n.contenido;
      const li = document.createElement("li");
      const info = document.createElement("div");
      const h = document.createElement("strong");
      h.textContent = n.titulo;
      const p = document.createElement("p");
      p.textContent = contenido || "";
      info.append(h, p);
      for (const c of n.categorias || []) info.append(chip(c.nombre));
      li.append(info, acciones("editar", "eliminar", n.id_nota));
      lista.appendChild(li);
    }
  } catch (e) {
    alert("Error al cargar notas: " + e.message);
  }
}

async function cargarCategorias() {
  try {
    const res = await fetch(API_CAT);
    if (!res.ok) throw new Error(res.status);
    const cats = (await res.json()) || [];
    categoriasCache = cats;
    listaCategorias.innerHTML = "";
    categoriasSelect.innerHTML = "";
    for (const c of cats) {
      const ch = document.createElement("li");
      ch.className = "chip";
      const nom = document.createElement("span");
      nom.textContent = c.nombre;
      ch.append(
        nom,
        boton("editar-cat", c.id_categoria, "✎"),
        boton("eliminar-cat", c.id_categoria, "×")
      );
      listaCategorias.appendChild(ch);
      const label = document.createElement("label");
      const input = document.createElement("input");
      input.type = "checkbox";
      input.value = c.id_categoria;
      const txt = document.createElement("span");
      txt.textContent = c.nombre;
      label.append(input, txt);
      categoriasSelect.appendChild(label);
    }
  } catch (e) {
    alert("Error al cargar categorías: " + e.message);
  }
}

$("btn-agregar-categoria").addEventListener("click", () => {
  categoriaForm.hidden = !categoriaForm.hidden;
  if (!categoriaForm.hidden) {
    editandoCategoria = null;
    categoriaForm.reset();
  }
});

$("btn-agregar-nota").addEventListener("click", () => {
  form.hidden = !form.hidden;
  if (!form.hidden) {
    editandoNota = null;
    form.reset();
  }
});

categoriaForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  try {
    const res = await fetch(
      editandoCategoria ? `${API_CAT}/${editandoCategoria}` : API_CAT,
      {
      method: editandoCategoria ? "PUT" : "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ nombre: categoriaForm.nombre.value }),
    });
    if (!res.ok) throw new Error(res.status);
    const edito = Boolean(editandoCategoria);
    editandoCategoria = null;
    categoriaForm.reset();
    cargarCategorias();
    if (edito) cargarNotas();
  } catch (e) {
    alert("Error al crear la categoría: " + e.message);
  }
});

form.addEventListener("submit", async (e) => {
  e.preventDefault();
  const categorias = Array.from(
    categoriasSelect.querySelectorAll("input[type=checkbox]:checked")
  ).map((i) => Number(i.value));
  try {
    const res = await fetch(editandoNota ? `${API}/${editandoNota}` : API, {
      method: editandoNota ? "PUT" : "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        titulo: form.titulo.value,
        contenido: form.contenido.value,
        is_archived: false,
        categorias,
      }),
    });
    if (!res.ok) throw new Error(res.status);
    editandoNota = null;
    form.reset();
    cargarNotas();
  } catch (e) {
    alert("Error al crear la nota: " + e.message);
  }
});

lista.addEventListener("click", async (e) => {
  const ed = e.target.closest(".editar");
  if (ed) {
    const n = notasCache.find((x) => String(x.id_nota) === String(ed.dataset.id));
    if (!n) return;
    editandoNota = n.id_nota;
    form.titulo.value = n.titulo;
    form.contenido.value =
      n.contenido && typeof n.contenido === "object" ? n.contenido.String : n.contenido || "";
    const ids = (n.categorias || []).map((c) => c.id_categoria);
    for (const i of categoriasSelect.querySelectorAll("input[type=checkbox]")) {
      i.checked = ids.includes(Number(i.value));
    }
    form.hidden = false;
    return;
  }
  const btn = e.target.closest(".eliminar");
  if (!btn) return;
  try {
    const res = await fetch(`${API}/${btn.dataset.id}`, { method: "DELETE" });
    if (!res.ok) throw new Error(res.status);
    cargarNotas();
  } catch (e) {
    alert("Error al eliminar: " + e.message);
  }
});

listaCategorias.addEventListener("click", async (e) => {
  const ed = e.target.closest(".editar-cat");
  if (ed) {
    const c = categoriasCache.find((x) => String(x.id_categoria) === String(ed.dataset.id));
    if (!c) return;
    editandoCategoria = c.id_categoria;
    categoriaForm.nombre.value = c.nombre;
    categoriaForm.hidden = false;
    return;
  }
  const del = e.target.closest(".eliminar-cat");
  if (!del) return;
  try {
    const res = await fetch(`${API_CAT}/${del.dataset.id}`, { method: "DELETE" });
    if (!res.ok) throw new Error(res.status);
    cargarCategorias();
    cargarNotas();
  } catch (e) {
    alert("Error al eliminar la categoría: " + e.message);
  }
});

document.addEventListener("DOMContentLoaded", () => {
  cargarCategorias();
  cargarNotas();
});
