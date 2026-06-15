/**
 * Products Module — Nexora Admin
 * Handles: CRUD, Images, Variants, Attributes
 */

// ============ STATE ============
const state = {
  currentPage: 1,
  limit: 20,
  search: '',
  status: '',
  category: '',
  editingProduct: null,
  selectedImages: [],       // File objects pending upload
  currentTab: 'general',
  categories: [],
  categoryAttributes: [],   // attributes for the selected category
  variantSelections: {},     // { attributeId: [selectedValues] }
  existingImages: [],        // images already stored on the server
  existingVariants: [],      // variants already stored on the server
};

// ============ ICONS (SVG) ============
const ICONS = {
  general: `<svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.2" viewBox="0 0 24 24" style="vertical-align:middle"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>`,
  precios: `<svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.2" viewBox="0 0 24 24" style="vertical-align:middle"><path stroke-linecap="round" stroke-linejoin="round" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>`,
  imagenes: `<svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.2" viewBox="0 0 24 24" style="vertical-align:middle"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>`,
  variantes: `<svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.2" viewBox="0 0 24 24" style="vertical-align:middle"><path stroke-linecap="round" stroke-linejoin="round" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"/></svg>`,
  box: `<svg width="40" height="40" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24" style="opacity:0.6"><path stroke-linecap="round" stroke-linejoin="round" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m-8-10l8 4m-8-4v10l8 4m0-14v4"/></svg>`,
  warning: `<svg width="40" height="40" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24" style="opacity:0.6"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/></svg>`,
  imageEmpty: `<svg width="40" height="40" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24" style="opacity:0.6"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>`,
  variantEmpty: `<svg width="40" height="40" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24" style="opacity:0.6"><path stroke-linecap="round" stroke-linejoin="round" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"/></svg>`,
  star: `<svg width="14" height="14" fill="#f59e0b" viewBox="0 0 24 24" style="vertical-align:middle"><path d="M12 17.27L18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"/></svg>`,
  camera: `<svg width="12" height="12" fill="none" stroke="currentColor" stroke-width="2.2" viewBox="0 0 24 24" style="display:inline-block;vertical-align:middle;margin-right:2px"><path stroke-linecap="round" stroke-linejoin="round" d="M6.827 6.175A2.31 2.31 0 015.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574V18a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 00-1.134-.175 2.31 2.31 0 01-1.64-1.055l-.822-1.316a2.192 2.192 0 00-1.736-1.039 48.774 48.774 0 00-5.232 0 2.192 2.192 0 00-1.736 1.039l-.821 1.316z"/><path stroke-linecap="round" stroke-linejoin="round" d="M16.5 12.75a4.5 4.5 0 11-9 0 4.5 4.5 0 019 0zM18.75 10.5h.008v.008h-.008V10.5z"/></svg>`,
  variantTiny: `<svg width="12" height="12" fill="none" stroke="currentColor" stroke-width="2.2" viewBox="0 0 24 24" style="display:inline-block;vertical-align:middle;margin-right:2px"><path stroke-linecap="round" stroke-linejoin="round" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"/></svg>`
};

// ============ FORMATTERS ============
const fmt = {
  price(v) {
    const n = Number(v) || 0;
    return n.toLocaleString('es-CO', { style: 'currency', currency: 'COP', minimumFractionDigits: 0 });
  },
  number(v) {
    return (Number(v) || 0).toLocaleString('es-CO');
  },
};

// ============ API HELPERS ============
async function api(url, options = {}) {
  const headers = { Authorization: `Bearer ${window.token}` };
  if (options.json) {
    headers['Content-Type'] = 'application/json';
    options.body = JSON.stringify(options.json);
    delete options.json;
  }
  const res = await fetch(`${window.API_URL}${url}`, { ...options, headers: { ...headers, ...options.headers } });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.message || `Error ${res.status}`);
  return data;
}

async function apiUpload(url, formData) {
  const res = await fetch(`${window.API_URL}${url}`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${window.token}` },
    body: formData,
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.message || `Error ${res.status}`);
  return data;
}

// ============ DOM HELPERS ============
function $(sel, parent = document) { return parent.querySelector(sel); }
function $$(sel, parent = document) { return [...parent.querySelectorAll(sel)]; }

function el(tag, attrs = {}, children = []) {
  const e = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs)) {
    if (k === 'className') e.className = v;
    else if (k === 'innerHTML') e.innerHTML = v;
    else if (k === 'textContent') e.textContent = v;
    else if (k.startsWith('on')) e.addEventListener(k.slice(2).toLowerCase(), v);
    else if (k === 'style' && typeof v === 'object') Object.assign(e.style, v);
    else e.setAttribute(k, v);
  }
  children.forEach(c => {
    if (typeof c === 'string') e.appendChild(document.createTextNode(c));
    else if (c) e.appendChild(c);
  });
  return e;
}

function setLoading(container, loading) {
  if (loading) {
    container.innerHTML = `
      <div style="display:flex;align-items:center;justify-content:center;padding:3rem;color:var(--text-sec)">
        <svg width="24" height="24" viewBox="0 0 24 24" style="animation:spin 1s linear infinite;margin-right:.5rem">
          <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" fill="none" stroke-dasharray="31" stroke-linecap="round"/>
        </svg>
        Cargando…
      </div>`;
  }
}

function emptyState(message, iconKey = 'box') {
  const svg = ICONS[iconKey] || ICONS['box'];
  return `
    <div style="display:flex;flex-direction:column;align-items:center;justify-content:center;padding:3rem;color:var(--text-mut)">
      <div style="margin-bottom:.75rem">${svg}</div>
      <span>${message}</span>
    </div>`;
}


// ============================================================
//                    PRODUCTS TABLE
// ============================================================
async function loadProducts(page = 1) {
  state.currentPage = page;
  const tbody = $('#products-tbody');
  if (!tbody) return;
  setLoading(tbody, true);

  try {
    const params = new URLSearchParams({
      page,
      limit: state.limit,
      search: state.search,
      estado: state.status,
      categoria: state.category,
    });
    const res = await api(`/products?${params}`);
    renderProducts(res.data || []);
    renderPagination(res.meta || { page: 1, total_pages: 1, total: 0 });
  } catch (err) {
    Toast.error('Error', err.message);
    tbody.innerHTML = emptyState('Error al cargar productos', 'warning');
  }
}

function renderProducts(products) {
  const tbody = $('#products-tbody');
  if (!products.length) {
    tbody.innerHTML = `<tr><td colspan="8">${emptyState('No se encontraron productos')}</td></tr>`;
    return;
  }

  tbody.innerHTML = products.map(p => {
    const img = p.imagenes && p.imagenes.length
      ? `<img src="${p.imagenes[0].url}" alt="" class="prod-thumb">`
      : `<div class="prod-thumb prod-thumb--placeholder"><svg width="20" height="20" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><rect x="3" y="3" width="18" height="18" rx="3"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="m21 15-5-5L5 21"/></svg></div>`;

    const stockClass = p.stock <= (p.stock_minimo || 5) ? 'badge badge-red' : 'badge badge-green';
    const estadoBadge = p.estado === 'activo'
      ? '<span class="badge badge-green">Activo</span>'
      : p.estado === 'borrador'
        ? '<span class="badge badge-yellow">Borrador</span>'
        : '<span class="badge badge-gray">' + (p.estado || '—') + '</span>';

    const variantBadge = p.tiene_variantes
      ? `<span class="badge badge-yellow" style="display:inline-flex;align-items:center;gap:4px">${ICONS.variantTiny} ${p.variantes_count || '—'}</span>`
      : '';
    const imgBadge = p.imagenes && p.imagenes.length
      ? `<span class="badge badge-gray" style="display:inline-flex;align-items:center;gap:4px">${ICONS.camera} ${p.imagenes.length}</span>`
      : '';

    return `
      <tr>
        <td>
          <div style="display:flex;align-items:center;gap:.75rem">
            ${img}
            <div>
              <div style="font-weight:600;color:var(--text)">${esc(p.nombre)}</div>
              <div style="font-size:.75rem;color:var(--text-mut)">${esc(p.sku || '—')}</div>
            </div>
          </div>
        </td>
        <td>${esc(p.categoria?.nombre || '—')}</td>
        <td style="font-weight:600">${fmt.price(p.precio)}</td>
        <td><span class="${stockClass}">${fmt.number(p.stock)}</span></td>
        <td>${variantBadge}</td>
        <td>${imgBadge}</td>
        <td>${estadoBadge}</td>
        <td>
          <div style="display:flex;gap:.35rem">
            <button class="action-btn" title="Editar" onclick="editProduct(${p.id})">
              <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
            </button>
            <button class="action-btn" title="Eliminar" onclick="deleteProduct(${p.id})">
              <svg width="16" height="16" fill="none" stroke="var(--danger)" stroke-width="2" viewBox="0 0 24 24"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
            </button>
          </div>
        </td>
      </tr>`;
  }).join('');
}

function esc(s) {
  if (s == null) return '';
  const d = document.createElement('div');
  d.textContent = String(s);
  return d.innerHTML;
}

// ============ PAGINATION ============
function renderPagination(meta) {
  const container = $('#pagination');
  if (!container) return;
  const { page, total_pages, total } = meta;
  if (total_pages <= 1) { container.innerHTML = ''; return; }

  let html = `<span style="color:var(--text-sec);font-size:.85rem">${fmt.number(total)} productos</span>`;
  html += `<div style="display:flex;gap:.25rem;align-items:center">`;

  // Previous
  html += `<button class="btn btn-sm btn-secondary" ${page <= 1 ? 'disabled' : ''} onclick="loadProducts(${page - 1})">
    <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><polyline points="15 18 9 12 15 6"/></svg>
  </button>`;

  // Page numbers
  const pages = buildPageNumbers(page, total_pages);
  pages.forEach(p => {
    if (p === '...') {
      html += `<span style="padding:0 .35rem;color:var(--text-mut)">…</span>`;
    } else {
      const active = p === page ? 'btn-primary' : 'btn-secondary';
      html += `<button class="btn btn-sm ${active}" onclick="loadProducts(${p})">${p}</button>`;
    }
  });

  // Next
  html += `<button class="btn btn-sm btn-secondary" ${page >= total_pages ? 'disabled' : ''} onclick="loadProducts(${page + 1})">
    <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><polyline points="9 18 15 12 9 6"/></svg>
  </button>`;

  html += `</div>`;
  container.innerHTML = html;
}

function buildPageNumbers(current, total) {
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1);
  const pages = [];
  pages.push(1);
  if (current > 3) pages.push('...');
  for (let i = Math.max(2, current - 1); i <= Math.min(total - 1, current + 1); i++) pages.push(i);
  if (current < total - 2) pages.push('...');
  pages.push(total);
  return pages;
}


// ============================================================
//                       CATEGORIES
// ============================================================
async function loadCategories() {
  try {
    const res = await api('/categories');
    state.categories = res.data || [];
    populateFilterCategories();
  } catch (err) {
    console.error('Error loading categories:', err);
  }
}

function populateFilterCategories() {
  const select = $('#filter-category');
  if (!select) return;
  select.innerHTML = '<option value="">Todas las categorías</option>';
  state.categories.forEach(c => {
    select.innerHTML += `<option value="${c.id}">${esc(c.nombre)}</option>`;
  });
}

function populateModalCategories(selectedId) {
  const select = $('#prod-categoria');
  if (!select) return;
  select.innerHTML = '<option value="">Seleccionar categoría</option>';
  state.categories.forEach(c => {
    const sel = c.id == selectedId ? 'selected' : '';
    select.innerHTML += `<option value="${c.id}" ${sel}>${esc(c.nombre)}</option>`;
  });
}


// ============================================================
//                  PRODUCT MODAL (TABS)
// ============================================================
function openModal(product = null) {
  state.editingProduct = product;
  state.selectedImages = [];
  state.existingImages = [];
  state.existingVariants = [];
  state.categoryAttributes = [];
  state.variantSelections = {};
  state.currentTab = 'general';

  const modal = $('#product-modal');
  if (!modal) { buildModal(); }

  const m = $('#product-modal');
  $('#modal-title').textContent = product ? 'Editar Producto' : 'Nuevo Producto';

  // Populate form
  populateModalCategories(product?.categoria_id);

  $('#prod-sku').value = product?.sku || '';
  $('#prod-nombre').value = product?.nombre || '';
  $('#prod-descripcion').value = product?.descripcion || '';
  $('#prod-estado').value = product?.estado || 'borrador';

  $('#prod-precio').value = product?.precio ?? '';
  $('#prod-precio-anterior').value = product?.precio_anterior ?? '';
  $('#prod-costo').value = product?.costo ?? '';
  $('#prod-stock').value = product?.stock ?? 0;
  $('#prod-stock-minimo').value = product?.stock_minimo ?? 5;

  // Reset tab & images
  switchTab('general');
  renderImagePreviews();
  renderExistingImages([]);
  renderVariantsTable([], product?.precio);
  renderAttributeSelectors([]);

  // If editing load images + variants + attributes
  if (product) {
    loadProductImages(product.id);
    loadVariants(product.id);
    if (product.categoria_id) loadCategoryAttributes(product.categoria_id);
  }

  m.classList.add('active');
}

function closeModal() {
  const m = $('#product-modal');
  if (m) m.classList.remove('active');
  state.editingProduct = null;
  state.selectedImages = [];
}

function switchTab(tabName) {
  state.currentTab = tabName;
  $$('.modal-tab-btn').forEach(b => b.classList.toggle('active', b.dataset.tab === tabName));
  $$('.modal-tab-panel').forEach(p => p.classList.toggle('active', p.dataset.tab === tabName));
}

function buildModal() {
  const overlay = el('div', { className: 'modal-overlay', id: 'product-modal', onClick(e) { if (e.target === this) closeModal(); } }, [
    el('div', { className: 'modal modal--product' }, [
      // Header
      el('div', { className: 'modal-header' }, [
        el('h2', { className: 'modal-title', id: 'modal-title', textContent: 'Producto' }),
        el('button', { className: 'modal-close', onClick: closeModal, innerHTML: '&times;' }),
      ]),

      // Tabs bar
      el('div', { className: 'modal-tabs', id: 'modal-tabs' }, [
        tabBtn('general', 'general', 'General'),
        tabBtn('precios', 'precios', 'Precios'),
        tabBtn('imagenes', 'imagenes', 'Imágenes'),
        tabBtn('variantes', 'variantes', 'Variantes'),
      ]),

      // Body
      el('div', { className: 'modal-body', id: 'modal-body' }, [
        buildTabGeneral(),
        buildTabPrecios(),
        buildTabImagenes(),
        buildTabVariantes(),
      ]),

      // Footer
      el('div', { className: 'modal-footer' }, [
        el('button', { className: 'btn btn-secondary', onClick: closeModal, textContent: 'Cancelar' }),
        el('button', { className: 'btn btn-primary', id: 'btn-save-product', onClick: saveProduct, textContent: 'Guardar' }),
      ]),
    ]),
  ]);
  document.body.appendChild(overlay);
}

function tabBtn(name, iconKey, label) {
  const active = name === 'general' ? ' active' : '';
  const svg = ICONS[iconKey] || '';
  return el('button', {
    className: `modal-tab-btn${active}`,
    'data-tab': name,
    onClick: () => switchTab(name),
  }, [
    el('span', { innerHTML: svg, style: { marginRight: '.35rem', display: 'inline-flex', alignItems: 'center' } }),
    el('span', { textContent: label }),
  ]);
}

// --- Tab: General ---
function buildTabGeneral() {
  const panel = el('div', { className: 'modal-tab-panel active', 'data-tab': 'general', id: 'tab-general' });
  panel.innerHTML = `
    <form id="product-form" onsubmit="return false">
      <div class="form-row">
        <div class="form-group" style="flex:1">
          <label class="form-label" for="prod-sku">SKU</label>
          <input type="text" id="prod-sku" class="form-input" placeholder="SKU-001" data-validate="required">
        </div>
        <div class="form-group" style="flex:2">
          <label class="form-label" for="prod-nombre">Nombre *</label>
          <input type="text" id="prod-nombre" class="form-input" placeholder="Nombre del producto" data-validate="required">
        </div>
      </div>
      <div class="form-group">
        <label class="form-label" for="prod-descripcion">Descripción</label>
        <textarea id="prod-descripcion" class="form-input" rows="3" placeholder="Descripción del producto…" style="resize:vertical"></textarea>
      </div>
      <div class="form-row">
        <div class="form-group" style="flex:1">
          <label class="form-label" for="prod-categoria">Categoría</label>
          <select id="prod-categoria" class="form-select" data-validate="required">
            <option value="">Seleccionar…</option>
          </select>
        </div>
        <div class="form-group" style="flex:1">
          <label class="form-label" for="prod-estado">Estado</label>
          <select id="prod-estado" class="form-select">
            <option value="borrador">Borrador</option>
            <option value="activo">Activo</option>
            <option value="inactivo">Inactivo</option>
          </select>
        </div>
      </div>
    </form>`;
  return panel;
}

// --- Tab: Precios ---
function buildTabPrecios() {
  const panel = el('div', { className: 'modal-tab-panel', 'data-tab': 'precios', id: 'tab-precios' });
  panel.innerHTML = `
    <div class="form-row">
      <div class="form-group" style="flex:1">
        <label class="form-label" for="prod-precio">Precio *</label>
        <input type="number" id="prod-precio" class="form-input" min="0" step="1" placeholder="0" data-validate="required|number">
      </div>
      <div class="form-group" style="flex:1">
        <label class="form-label" for="prod-precio-anterior">Precio anterior</label>
        <input type="number" id="prod-precio-anterior" class="form-input" min="0" step="1" placeholder="0">
      </div>
    </div>
    <div class="form-group">
      <label class="form-label" for="prod-costo">Costo</label>
      <input type="number" id="prod-costo" class="form-input" min="0" step="1" placeholder="0">
    </div>
    <div class="form-row">
      <div class="form-group" style="flex:1">
        <label class="form-label" for="prod-stock">Stock</label>
        <input type="number" id="prod-stock" class="form-input" min="0" step="1" value="0">
      </div>
      <div class="form-group" style="flex:1">
        <label class="form-label" for="prod-stock-minimo">Stock mínimo</label>
        <input type="number" id="prod-stock-minimo" class="form-input" min="0" step="1" value="5">
      </div>
    </div>`;
  return panel;
}

// --- Tab: Imágenes ---
function buildTabImagenes() {
  const panel = el('div', { className: 'modal-tab-panel', 'data-tab': 'imagenes', id: 'tab-imagenes' });
  panel.innerHTML = `
    <div class="img-upload-zone" id="img-upload-zone">
      <input type="file" id="img-file-input" accept="image/*" multiple hidden>
      <div class="img-upload-zone__content">
        <svg width="40" height="40" fill="none" stroke="var(--text-mut)" stroke-width="1.5" viewBox="0 0 24 24">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="17 8 12 3 7 8"/>
          <line x1="12" y1="3" x2="12" y2="15"/>
        </svg>
        <p style="margin:.5rem 0 0;color:var(--text-sec)">Arrastra imágenes aquí o <span style="color:var(--accent);cursor:pointer;text-decoration:underline">selecciona archivos</span></p>
        <span style="font-size:.75rem;color:var(--text-mut)">PNG, JPG, WEBP — máx. 5 MB</span>
      </div>
    </div>
    <div id="img-previews" class="img-preview-grid" style="margin-top:1rem"></div>
    <div id="img-upload-actions" style="display:none;margin-top:.75rem;text-align:right">
      <button class="btn btn-sm btn-secondary" onclick="clearSelectedImages()">Limpiar</button>
      <button class="btn btn-sm btn-primary" onclick="uploadSelectedImages()" style="margin-left:.5rem">
        <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" style="margin-right:.25rem"><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
        Subir
      </button>
    </div>
    <div style="margin-top:1.25rem">
      <h4 style="color:var(--text);margin:0 0 .75rem;display:flex;align-items:center;gap:.5rem">
        Imágenes del producto <span id="img-count-badge" class="badge badge-gray" style="font-size:.7rem">0 fotos</span>
      </h4>
      <div id="img-existing" class="img-existing-grid"></div>
    </div>`;
  return panel;
}

// --- Tab: Variantes ---
function buildTabVariantes() {
  const panel = el('div', { className: 'modal-tab-panel', 'data-tab': 'variantes', id: 'tab-variantes' });
  panel.innerHTML = `
    <div id="variant-attrs-section">
      <h4 style="color:var(--text);margin:0 0 .75rem">Atributos de la categoría</h4>
      <div id="variant-attrs-container" style="margin-bottom:1rem"></div>
      <button class="btn btn-sm btn-primary" id="btn-generate-variants" onclick="handleGenerateVariants()" style="margin-bottom:1.25rem">
        <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" style="margin-right:.25rem"><path d="M12 5v14M5 12h14"/></svg>
        Generar Variantes
      </button>
    </div>
    <div id="variants-table-wrapper"></div>
    <div id="variant-actions" style="display:none;margin-top:.75rem;display:flex;justify-content:flex-end;gap:.5rem">
      <button class="btn btn-sm btn-primary" id="btn-bulk-save" onclick="handleBulkSave()">
        <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" style="margin-right:.25rem"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/></svg>
        Guardar todo
      </button>
    </div>`;
  return panel;
}


// ============================================================
//                     SAVE / EDIT / DELETE
// ============================================================
async function saveProduct() {
  // Validate general + prices tabs
  if (!FormValidator.validateForm('product-form')) {
    switchTab('general');
    return;
  }

  const body = {
    sku: $('#prod-sku').value.trim(),
    nombre: $('#prod-nombre').value.trim(),
    descripcion: $('#prod-descripcion').value.trim(),
    categoria_id: Number($('#prod-categoria').value) || null,
    estado: $('#prod-estado').value,
    precio: Number($('#prod-precio').value) || 0,
    precio_anterior: Number($('#prod-precio-anterior').value) || null,
    costo: Number($('#prod-costo').value) || null,
    stock: Number($('#prod-stock').value) || 0,
    stock_minimo: Number($('#prod-stock-minimo').value) || 5,
  };

  const btn = $('#btn-save-product');
  btn.disabled = true;
  btn.textContent = 'Guardando…';

  try {
    let res;
    if (state.editingProduct) {
      res = await api(`/products/${state.editingProduct.id}`, { method: 'PUT', json: body });
    } else {
      res = await api('/products', { method: 'POST', json: body });
    }

    const savedId = res.data?.id || state.editingProduct?.id;

    // Upload pending images if any
    if (state.selectedImages.length && savedId) {
      await uploadImages(savedId);
    }

    Toast.success('Éxito', res.message || 'Producto guardado correctamente');
    closeModal();
    loadProducts(state.currentPage);
  } catch (err) {
    Toast.error('Error', err.message);
  } finally {
    btn.disabled = false;
    btn.textContent = 'Guardar';
  }
}

async function editProduct(id) {
  try {
    const res = await api(`/products/${id}`);
    openModal(res.data);
  } catch (err) {
    Toast.error('Error', err.message);
  }
}

async function deleteProduct(id) {
  const confirmed = await window.confirmDelete({
    title: 'Eliminar producto',
    message: '¿Estás seguro de que deseas eliminar este producto? Esta acción no se puede deshacer.',
  });
  if (!confirmed) return;

  try {
    const res = await api(`/products/${id}`, { method: 'DELETE' });
    Toast.success('Eliminado', res.message || 'Producto eliminado');
    loadProducts(state.currentPage);
  } catch (err) {
    Toast.error('Error', err.message);
  }
}


// ============================================================
//                     IMAGE MANAGEMENT
// ============================================================
function initImageUpload() {
  // We defer setup until the modal DOM exists
  const zone = $('#img-upload-zone');
  if (!zone) return;

  const input = $('#img-file-input');

  // Click to select
  zone.addEventListener('click', () => input.click());

  // File input change
  input.addEventListener('change', () => {
    if (input.files.length) handleImageSelect(input.files);
    input.value = '';
  });

  // Drag & drop
  zone.addEventListener('dragover', e => { e.preventDefault(); zone.classList.add('dragover'); });
  zone.addEventListener('dragleave', () => zone.classList.remove('dragover'));
  zone.addEventListener('drop', e => {
    e.preventDefault();
    zone.classList.remove('dragover');
    if (e.dataTransfer.files.length) handleImageSelect(e.dataTransfer.files);
  });
}

function handleImageSelect(fileList) {
  const files = [...fileList].filter(f => f.type.startsWith('image/'));
  if (!files.length) return;
  const maxSize = 5 * 1024 * 1024;
  for (const f of files) {
    if (f.size > maxSize) {
      Toast.error('Archivo grande', `"${f.name}" supera el límite de 5 MB`);
      continue;
    }
    state.selectedImages.push(f);
  }
  renderImagePreviews();
}

function renderImagePreviews() {
  const container = $('#img-previews');
  const actions = $('#img-upload-actions');
  if (!container) return;

  if (!state.selectedImages.length) {
    container.innerHTML = '';
    if (actions) actions.style.display = 'none';
    return;
  }

  if (actions) actions.style.display = 'block';

  container.innerHTML = state.selectedImages.map((f, i) => {
    const url = URL.createObjectURL(f);
    return `
      <div class="img-preview-item">
        <img src="${url}" alt="">
        <button class="img-preview-remove" onclick="removeSelectedImage(${i})" title="Quitar">&times;</button>
        <span class="img-preview-name">${esc(f.name)}</span>
      </div>`;
  }).join('');
}

function removeSelectedImage(index) {
  state.selectedImages.splice(index, 1);
  renderImagePreviews();
}

function clearSelectedImages() {
  state.selectedImages = [];
  renderImagePreviews();
}

async function uploadSelectedImages() {
  if (!state.editingProduct) {
    Toast.error('Aviso', 'Guarda el producto primero antes de subir imágenes');
    return;
  }
  if (!state.selectedImages.length) return;
  await uploadImages(state.editingProduct.id);
}

async function uploadImages(productId) {
  if (!state.selectedImages.length) return;
  const formData = new FormData();
  state.selectedImages.forEach(f => formData.append('images', f));

  try {
    await apiUpload(`/products/${productId}/images`, formData);
    Toast.success('Imágenes', 'Imágenes subidas correctamente');
    state.selectedImages = [];
    renderImagePreviews();
    loadProductImages(productId);
  } catch (err) {
    Toast.error('Error', err.message);
  }
}

async function loadProductImages(productId) {
  const container = $('#img-existing');
  if (!container) return;
  setLoading(container, true);

  try {
    const res = await api(`/products/${productId}/images`);
    state.existingImages = res.data || [];
    renderExistingImages(state.existingImages);
  } catch (err) {
    container.innerHTML = emptyState('Error al cargar imágenes', 'warning');
  }
}

function renderExistingImages(images) {
  const container = $('#img-existing');
  const badge = $('#img-count-badge');
  if (!container) return;
  if (badge) badge.textContent = `${images.length} foto${images.length !== 1 ? 's' : ''}`;

  if (!images.length) {
    container.innerHTML = emptyState('Sin imágenes', 'imageEmpty');
    return;
  }

  container.innerHTML = images.map(img => {
    const isPrincipal = img.es_principal || img.principal;
    return `
      <div class="img-existing-item ${isPrincipal ? 'img-existing-item--principal' : ''}">
        <img src="${img.url}" alt="" loading="lazy">
        ${isPrincipal ? `<div class="img-star-badge" title="Imagen principal">${ICONS.star}</div>` : ''}
        <div class="img-hover-overlay">
          ${!isPrincipal ? `<button class="img-action-btn" onclick="handleSetPrincipal(${img.id})" title="Establecer como principal">
            <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
          </button>` : ''}
          <button class="img-action-btn img-action-btn--danger" onclick="handleDeleteImage(${img.id})" title="Eliminar">
            <svg width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
          </button>
        </div>
      </div>`;
  }).join('');
}

async function handleSetPrincipal(imageId) {
  if (!state.editingProduct) return;
  try {
    await api(`/products/${state.editingProduct.id}/images/${imageId}/principal`, { method: 'PUT' });
    Toast.success('Imagen', 'Imagen principal actualizada');
    loadProductImages(state.editingProduct.id);
  } catch (err) {
    Toast.error('Error', err.message);
  }
}

async function handleDeleteImage(imageId) {
  if (!state.editingProduct) return;
  const confirmed = await window.confirmDelete({
    title: 'Eliminar imagen',
    message: '¿Estás seguro de que deseas eliminar esta imagen?',
  });
  if (!confirmed) return;

  try {
    await api(`/products/${state.editingProduct.id}/images/${imageId}`, { method: 'DELETE' });
    Toast.success('Imagen', 'Imagen eliminada');
    loadProductImages(state.editingProduct.id);
  } catch (err) {
    Toast.error('Error', err.message);
  }
}


// ============================================================
//                   VARIANT MANAGEMENT
// ============================================================
async function loadCategoryAttributes(categoryId) {
  if (!categoryId) {
    state.categoryAttributes = [];
    state.variantSelections = {};
    renderAttributeSelectors([]);
    return;
  }

  try {
    // Fetch category-specific + global attributes in parallel
    const [catRes, globalRes] = await Promise.all([
      api(`/attributes/category/${categoryId}`),
      api('/attributes/global'),
    ]);

    const catAttrs = catRes.data || [];
    const globalAttrs = globalRes.data || [];

    // Merge, avoid duplicates by id
    const seen = new Set(catAttrs.map(a => a.id));
    const merged = [...catAttrs];
    globalAttrs.forEach(a => { if (!seen.has(a.id)) merged.push(a); });

    state.categoryAttributes = merged;
    state.variantSelections = {};
    renderAttributeSelectors(merged);
  } catch (err) {
    console.error('Error loading attributes:', err);
    renderAttributeSelectors([]);
  }
}

function renderAttributeSelectors(attributes) {
  const container = $('#variant-attrs-container');
  if (!container) return;

  if (!attributes.length) {
    container.innerHTML = `<p style="color:var(--text-mut);font-size:.85rem">Selecciona una categoría para ver los atributos disponibles.</p>`;
    return;
  }

  container.innerHTML = attributes.map(attr => {
    const values = attr.valores || attr.values || [];
    if (!values.length) return '';

    const chips = values.map(v => {
      const valStr = typeof v === 'string' ? v : v.valor || v.value || v.nombre || '';
      const selected = (state.variantSelections[attr.id] || []).includes(valStr);
      return `<button type="button" class="attr-chip ${selected ? 'attr-chip--selected' : ''}"
                onclick="toggleAttributeValue(${attr.id}, '${esc(valStr)}')">${esc(valStr)}</button>`;
    }).join('');

    return `
      <div class="attr-group">
        <label class="form-label" style="margin-bottom:.35rem">${esc(attr.nombre || attr.name)}</label>
        <div class="attr-chips">${chips}</div>
      </div>`;
  }).join('');
}

function toggleAttributeValue(attrId, value) {
  if (!state.variantSelections[attrId]) state.variantSelections[attrId] = [];
  const arr = state.variantSelections[attrId];
  const idx = arr.indexOf(value);
  if (idx >= 0) arr.splice(idx, 1);
  else arr.push(value);
  renderAttributeSelectors(state.categoryAttributes);
}

async function handleGenerateVariants() {
  if (!state.editingProduct) {
    Toast.error('Aviso', 'Guarda el producto primero antes de generar variantes');
    return;
  }

  // Build the selections payload
  const atributos = Object.entries(state.variantSelections)
    .filter(([, vals]) => vals.length > 0)
    .map(([id, valores]) => ({ atributo_id: Number(id), valores }));

  if (!atributos.length) {
    Toast.error('Aviso', 'Selecciona al menos un valor de atributo');
    return;
  }

  const btn = $('#btn-generate-variants');
  btn.disabled = true;
  btn.textContent = 'Generando…';

  try {
    const res = await api(`/products/${state.editingProduct.id}/variants/generate`, {
      method: 'POST',
      json: { atributos },
    });
    Toast.success('Variantes', res.message || 'Variantes generadas correctamente');
    loadVariants(state.editingProduct.id);
  } catch (err) {
    Toast.error('Error', err.message);
  } finally {
    btn.disabled = false;
    btn.innerHTML = `<svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" style="margin-right:.25rem"><path d="M12 5v14M5 12h14"/></svg> Generar Variantes`;
  }
}

async function loadVariants(productId) {
  const wrapper = $('#variants-table-wrapper');
  if (!wrapper) return;
  setLoading(wrapper, true);

  try {
    const res = await api(`/products/${productId}/variants`);
    state.existingVariants = res.data || [];
    const basePrice = state.editingProduct?.precio || 0;
    renderVariantsTable(state.existingVariants, basePrice);
  } catch (err) {
    wrapper.innerHTML = emptyState('Error al cargar variantes', 'warning');
  }
}

function renderVariantsTable(variants, basePrice) {
  const wrapper = $('#variants-table-wrapper');
  const actionsDiv = $('#variant-actions');
  if (!wrapper) return;

  if (!variants.length) {
    wrapper.innerHTML = emptyState('Sin variantes creadas', 'variantEmpty');
    if (actionsDiv) actionsDiv.style.display = 'none';
    return;
  }

  if (actionsDiv) actionsDiv.style.display = 'flex';

  let totalStock = 0;
  const rows = variants.map((v, i) => {
    const attrChips = (v.atributos || []).map(a => {
      const color = chipColor(a.atributo_nombre || a.nombre || '');
      return `<span class="variant-chip" style="background:${color}">${esc(a.valor || a.value || '')}</span>`;
    }).join('');

    const stock = Number(v.stock) || 0;
    totalStock += stock;
    const active = v.activo !== false;

    return `
      <tr data-variant-id="${v.id}" data-index="${i}">
        <td>${attrChips || '—'}</td>
        <td>
          <input type="number" class="form-input form-input--sm var-precio"
                 value="${v.precio ?? basePrice ?? ''}" min="0" step="1"
                 style="width:100px">
        </td>
        <td>
          <input type="number" class="form-input form-input--sm var-stock"
                 value="${stock}" min="0" step="1" style="width:80px">
        </td>
        <td>
          <label class="toggle-switch">
            <input type="checkbox" class="var-activo" ${active ? 'checked' : ''}>
            <span class="toggle-slider"></span>
          </label>
        </td>
        <td>
          <button class="action-btn" title="Eliminar variante" onclick="handleDeleteVariant(${v.id})">
            <svg width="14" height="14" fill="none" stroke="var(--danger)" stroke-width="2" viewBox="0 0 24 24"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
          </button>
        </td>
      </tr>`;
  }).join('');

  wrapper.innerHTML = `
    <div style="overflow-x:auto">
      <table class="variants-table" id="variants-table">
        <thead>
          <tr>
            <th>Atributos</th>
            <th>Precio</th>
            <th>Stock</th>
            <th>Activo</th>
            <th></th>
          </tr>
        </thead>
        <tbody>${rows}</tbody>
        <tfoot>
          <tr>
            <td style="font-weight:600;color:var(--text)">Total</td>
            <td></td>
            <td style="font-weight:600;color:var(--accent)">${fmt.number(totalStock)}</td>
            <td></td>
            <td></td>
          </tr>
        </tfoot>
      </table>
    </div>`;
}

/** Generate a pastel background color deterministically from a string */
function chipColor(str) {
  let hash = 0;
  for (let i = 0; i < str.length; i++) hash = str.charCodeAt(i) + ((hash << 5) - hash);
  const h = Math.abs(hash) % 360;
  return `hsl(${h}, 55%, 25%)`;
}

async function handleBulkSave() {
  if (!state.editingProduct) return;
  const rows = $$('#variants-table tbody tr');
  if (!rows.length) return;

  const variantes = rows.map(row => ({
    id: Number(row.dataset.variantId),
    precio: Number($('.var-precio', row).value) || 0,
    stock: Number($('.var-stock', row).value) || 0,
    activo: $('.var-activo', row).checked,
  }));

  const btn = $('#btn-bulk-save');
  btn.disabled = true;
  btn.textContent = 'Guardando…';

  try {
    await api(`/products/${state.editingProduct.id}/variants/bulk`, {
      method: 'PUT',
      json: { variantes },
    });
    Toast.success('Variantes', 'Variantes actualizadas correctamente');
    loadVariants(state.editingProduct.id);
  } catch (err) {
    Toast.error('Error', err.message);
  } finally {
    btn.disabled = false;
    btn.innerHTML = `<svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" style="margin-right:.25rem"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/></svg> Guardar todo`;
  }
}

async function handleDeleteVariant(variantId) {
  if (!state.editingProduct) return;
  const confirmed = await window.confirmDelete({
    title: 'Eliminar variante',
    message: '¿Eliminar esta variante?',
  });
  if (!confirmed) return;

  try {
    await api(`/products/${state.editingProduct.id}/variants/${variantId}`, { method: 'DELETE' });
    Toast.success('Variante', 'Variante eliminada');
    loadVariants(state.editingProduct.id);
  } catch (err) {
    Toast.error('Error', err.message);
  }
}


// ============================================================
//                     INLINE STYLES
// ============================================================
function injectStyles() {
  if ($('#products-module-styles')) return;
  const style = document.createElement('style');
  style.id = 'products-module-styles';
  style.textContent = `
    /* ---- Product thumbnail ---- */
    .prod-thumb {
      width: 44px; height: 44px; border-radius: 8px; object-fit: cover;
      border: 1px solid var(--border); flex-shrink: 0;
    }
    .prod-thumb--placeholder {
      display: flex; align-items: center; justify-content: center;
      background: var(--bg-surf); color: var(--text-mut);
    }

    /* ---- Modal sizing ---- */
    .modal--product { max-width: 720px; width: 95vw; }

    /* ---- Tabs ---- */
    .modal-tabs {
      display: flex; gap: 0; border-bottom: 1px solid var(--border);
      padding: 0 1.25rem; overflow-x: auto;
    }
    .modal-tab-btn {
      background: none; border: none; color: var(--text-sec); cursor: pointer;
      padding: .65rem 1rem; font-size: .85rem; font-weight: 500;
      border-bottom: 2px solid transparent; transition: all .2s;
      white-space: nowrap; display: flex; align-items: center;
    }
    .modal-tab-btn:hover { color: var(--text); }
    .modal-tab-btn.active {
      color: var(--accent); border-bottom-color: var(--accent);
    }
    .modal-tab-panel { display: none; }
    .modal-tab-panel.active { display: block; }

    /* ---- Form helpers ---- */
    .form-input--sm { padding: .35rem .5rem; font-size: .8rem; }

    /* ---- Image upload zone ---- */
    .img-upload-zone {
      border: 2px dashed var(--border); border-radius: 12px; padding: 2rem;
      text-align: center; cursor: pointer; transition: all .25s;
    }
    .img-upload-zone:hover,
    .img-upload-zone.dragover {
      border-color: var(--accent); background: var(--accent-dim);
    }

    /* ---- Image preview grid ---- */
    .img-preview-grid {
      display: grid; grid-template-columns: repeat(auto-fill, minmax(100px, 1fr)); gap: .75rem;
    }
    .img-preview-item {
      position: relative; border-radius: 8px; overflow: hidden;
      aspect-ratio: 1; border: 1px solid var(--border);
    }
    .img-preview-item img {
      width: 100%; height: 100%; object-fit: cover;
    }
    .img-preview-remove {
      position: absolute; top: 4px; right: 4px; width: 22px; height: 22px;
      border-radius: 50%; background: rgba(0,0,0,.7); color: #fff; border: none;
      cursor: pointer; font-size: 14px; line-height: 1; display: flex;
      align-items: center; justify-content: center;
    }
    .img-preview-name {
      position: absolute; bottom: 0; left: 0; right: 0; padding: 2px 4px;
      font-size: .65rem; color: #fff; background: rgba(0,0,0,.6);
      white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
    }

    /* ---- Existing images grid ---- */
    .img-existing-grid {
      display: grid; grid-template-columns: repeat(3, 1fr); gap: .75rem;
    }
    @media (max-width: 600px) {
      .img-existing-grid { grid-template-columns: repeat(2, 1fr); }
    }
    .img-existing-item {
      position: relative; border-radius: 10px; overflow: hidden;
      aspect-ratio: 1; border: 2px solid var(--border); transition: border-color .2s;
    }
    .img-existing-item--principal { border-color: var(--warning); }
    .img-existing-item img {
      width: 100%; height: 100%; object-fit: cover;
    }
    .img-star-badge {
      position: absolute; top: 6px; left: 6px; font-size: .9rem;
      filter: drop-shadow(0 1px 2px rgba(0,0,0,.5));
    }
    .img-hover-overlay {
      position: absolute; inset: 0; display: flex; align-items: center;
      justify-content: center; gap: .5rem; background: rgba(0,0,0,.55);
      opacity: 0; transition: opacity .2s;
    }
    .img-existing-item:hover .img-hover-overlay { opacity: 1; }
    /* Touch: keep overlay visible on mobile */
    @media (hover: none) {
      .img-hover-overlay { opacity: 1; background: rgba(0,0,0,.35); }
    }
    .img-action-btn {
      width: 36px; height: 36px; border-radius: 50%; border: none;
      background: rgba(255,255,255,.2); color: #fff; cursor: pointer;
      display: flex; align-items: center; justify-content: center;
      backdrop-filter: blur(4px); transition: background .2s;
    }
    .img-action-btn:hover { background: rgba(255,255,255,.35); }
    .img-action-btn--danger:hover { background: rgba(220,38,38,.7); }

    /* ---- Attribute chips ---- */
    .attr-group { margin-bottom: .75rem; }
    .attr-chips { display: flex; flex-wrap: wrap; gap: .4rem; }
    .attr-chip {
      padding: .3rem .7rem; border-radius: 20px; font-size: .8rem;
      border: 1px solid var(--border); background: var(--bg-surf);
      color: var(--text-sec); cursor: pointer; transition: all .2s;
    }
    .attr-chip:hover { border-color: var(--accent); color: var(--text); }
    .attr-chip--selected {
      background: var(--accent); border-color: var(--accent);
      color: #fff; font-weight: 600;
    }

    /* ---- Variants table ---- */
    .variants-table {
      width: 100%; border-collapse: separate; border-spacing: 0;
      font-size: .85rem;
    }
    .variants-table th {
      text-align: left; padding: .5rem .6rem; color: var(--text-mut);
      font-weight: 600; font-size: .75rem; text-transform: uppercase;
      letter-spacing: .04em; border-bottom: 1px solid var(--border);
    }
    .variants-table td {
      padding: .45rem .6rem; border-bottom: 1px solid var(--border);
      vertical-align: middle;
    }
    .variants-table tfoot td {
      border-bottom: none; padding-top: .65rem;
    }
    .variant-chip {
      display: inline-block; padding: .15rem .55rem; border-radius: 12px;
      font-size: .75rem; font-weight: 500; color: #e2e8f0; margin-right: .25rem;
    }

    /* ---- Toggle switch ---- */
    .toggle-switch {
      position: relative; display: inline-block; width: 36px; height: 20px;
    }
    .toggle-switch input { opacity: 0; width: 0; height: 0; }
    .toggle-slider {
      position: absolute; cursor: pointer; inset: 0;
      background: var(--border); border-radius: 20px; transition: .25s;
    }
    .toggle-slider::before {
      content: ''; position: absolute; width: 14px; height: 14px;
      left: 3px; bottom: 3px; background: #fff; border-radius: 50%;
      transition: .25s;
    }
    .toggle-switch input:checked + .toggle-slider { background: var(--accent); }
    .toggle-switch input:checked + .toggle-slider::before { transform: translateX(16px); }

    /* ---- Spin animation ---- */
    @keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
  `;
  document.head.appendChild(style);
}


// ============================================================
//                     INITIALIZATION
// ============================================================
document.addEventListener('DOMContentLoaded', () => {
  injectStyles();

  // Build the modal once
  buildModal();
  initImageUpload();

  // Load initial data
  loadCategories();
  loadProducts();

  // --- Search ---
  const searchInput = $('#search-input');
  if (searchInput) {
    let debounce;
    searchInput.addEventListener('input', () => {
      clearTimeout(debounce);
      debounce = setTimeout(() => {
        state.search = searchInput.value.trim();
        loadProducts(1);
      }, 350);
    });
  }

  // --- Filters ---
  const filterStatus = $('#filter-status');
  if (filterStatus) {
    filterStatus.addEventListener('change', () => {
      state.status = filterStatus.value;
      loadProducts(1);
    });
  }

  const filterCategory = $('#filter-category');
  if (filterCategory) {
    filterCategory.addEventListener('change', () => {
      state.category = filterCategory.value;
      loadProducts(1);
    });
  }

  // --- New product button ---
  const btnNew = $('#btn-new-product');
  if (btnNew) {
    btnNew.addEventListener('click', () => openModal());
  }

  // --- Category change inside modal → load attributes ---
  // We use event delegation since the modal is created dynamically
  document.addEventListener('change', e => {
    if (e.target && e.target.id === 'prod-categoria') {
      const catId = e.target.value;
      loadCategoryAttributes(catId);
    }
  });
});
