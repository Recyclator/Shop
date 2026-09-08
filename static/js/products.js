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
  variantAttributes: [],    // active attributes for generation: [{ id, name, values: [], suggestions: [] }]
  variantSelections: {},     // { attributeId: [selectedValues] }
  existingImages: [],        // images already stored on the server
  existingVariants: [],      // variants list (both existing and pending in memory)
  variantPendingFiles: {},   // { variantKey: [File] }
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
  state.variantPendingFiles = {};
  state.variantAttributes = [];
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
  $('#prod-estado').value = product?.estado || 'activo';

  $('#prod-precio').value = product?.precio ?? '';
  $('#prod-precio-anterior').value = product?.precio_anterior ?? '';
  $('#prod-costo').value = product?.costo ?? '';
  $('#prod-stock').value = product?.stock ?? 0;
  $('#prod-stock-minimo').value = product?.stock_minimo ?? 5;

  renderImagePreviews();
  renderExistingImages([]);
  renderAttributeBuilder();
  renderVariantsTable([], product?.precio);

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
  state.existingVariants = [];
  state.variantPendingFiles = {};
  state.variantAttributes = [];
}

function switchTab(tabName) {
  // Tabs removed, unified layout used
}

function buildModal() {
  const overlay = el('div', { className: 'modal-overlay', id: 'product-modal', onClick(e) { if (e.target === this) closeModal(); } }, [
    el('div', { className: 'modal modal--product' }, [
      // Header
      el('div', { className: 'modal-header' }, [
        el('h2', { className: 'modal-title', id: 'modal-title', textContent: 'Producto' }),
        el('button', { className: 'modal-close', onClick: closeModal, innerHTML: '&times;', 'aria-label': 'Cerrar' }),
      ]),

      // Body
      el('div', { className: 'modal-body', id: 'modal-body' }),

      // Footer
      el('div', { className: 'modal-footer' }, [
        el('button', { className: 'btn btn-secondary', onClick: closeModal, textContent: 'Cancelar' }),
        el('button', { className: 'btn btn-primary', id: 'btn-save-product', onClick: saveProduct, textContent: 'Guardar' }),
      ]),
    ]),
  ]);
  
  const body = overlay.querySelector('#modal-body');
  body.innerHTML = `
    <form id="product-form" onsubmit="return false" class="prod-modal-form">
      
      <!-- Sección 1: Información General y Precios -->
      <div class="prod-form-section">
        <div class="prod-section-header">
          <span class="prod-section-title">
            <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>
            Información General
          </span>
        </div>
        
        <div class="prod-grid prod-grid--header">
          <div class="form-group prod-col--sku">
            <label class="form-label" for="prod-sku">SKU *</label>
            <input type="text" id="prod-sku" class="form-input" placeholder="ej. CAM-001" required data-validate="required">
          </div>
          <div class="form-group prod-col--name">
            <label class="form-label" for="prod-nombre">Nombre *</label>
            <input type="text" id="prod-nombre" class="form-input" placeholder="ej. Camiseta Polo Premium" required data-validate="required">
          </div>
          <div class="form-group prod-col--cat">
            <label class="form-label" for="prod-categoria">Categoría *</label>
            <select id="prod-categoria" class="form-select" required data-validate="required">
              <option value="">Seleccionar…</option>
            </select>
          </div>
          <div class="form-group prod-col--status">
            <label class="form-label" for="prod-estado">Estado</label>
            <select id="prod-estado" class="form-select">
              <option value="activo">Activo</option>
              <option value="borrador">Borrador</option>
              <option value="inactivo">Inactivo</option>
            </select>
          </div>
        </div>

        <div class="form-group" style="margin:0;">
          <label class="form-label" for="prod-descripcion">Descripción</label>
          <textarea id="prod-descripcion" class="form-input prod-textarea" rows="2" placeholder="Detalles, material o especificaciones del producto…"></textarea>
        </div>

        <div class="prod-grid prod-grid--pricing">
          <div class="form-group">
            <label class="form-label" for="prod-precio">Precio venta base *</label>
            <input type="number" id="prod-precio" class="form-input" min="0" step="1" placeholder="0" required data-validate="required|number">
          </div>
          <div class="form-group">
            <label class="form-label" for="prod-precio-anterior">Precio anterior</label>
            <input type="number" id="prod-precio-anterior" class="form-input" min="0" step="1" placeholder="0">
          </div>
          <div class="form-group">
            <label class="form-label" for="prod-costo">Costo</label>
            <input type="number" id="prod-costo" class="form-input" min="0" step="1" placeholder="0">
          </div>
          <div class="form-group">
            <label class="form-label" for="prod-stock-minimo">Stock mínimo</label>
            <input type="number" id="prod-stock-minimo" class="form-input" min="0" step="1" value="5">
          </div>
          <div class="form-group">
            <label class="form-label" for="prod-stock" title="Calculado automáticamente al agregar variantes">Stock total</label>
            <input type="number" id="prod-stock" class="form-input" min="0" step="1" value="0">
          </div>
        </div>
      </div>

      <!-- Sección 2: Galería / Imágenes Principales -->
      <div class="prod-form-section">
        <div class="prod-section-header">
          <span class="prod-section-title">
            <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>
            Galería General del Producto
          </span>
          <span id="img-count-badge" class="badge badge-gray" style="font-size:.7rem">0 fotos</span>
        </div>
        
        <div class="prod-img-layout">
          <div class="img-upload-zone" id="img-upload-zone">
            <input type="file" id="img-file-input" accept="image/*" multiple hidden>
            <div class="img-upload-zone__content">
              <svg width="22" height="22" fill="none" stroke="var(--accent)" stroke-width="2" viewBox="0 0 24 24">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                <polyline points="17 8 12 3 7 8"/>
                <line x1="12" y1="3" x2="12" y2="15"/>
              </svg>
              <p class="img-upload-zone__text">Arrastra imágenes aquí o <span class="img-upload-link">examinar</span></p>
              <span class="img-upload-zone__hint">PNG, JPG, WEBP — máx. 5 MB</span>
            </div>
          </div>

          <div class="prod-img-content">
            <div id="img-previews" class="img-preview-grid"></div>
            <div id="img-upload-actions" style="display:none; margin-bottom:.5rem; text-align:right">
              <button type="button" class="btn btn-sm btn-secondary" onclick="clearSelectedImages()">Limpiar</button>
              <button type="button" id="btn-upload-photos-direct" class="btn btn-sm btn-primary" onclick="uploadSelectedImages()" style="margin-left:.4rem">
                <svg width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" style="margin-right:.25rem"><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                Subir fotos
              </button>
            </div>
            <div id="img-existing" class="img-existing-grid"></div>
          </div>
        </div>
      </div>

      <!-- Sección 3: Variantes -->
      <div class="prod-form-section" id="modal-section-variants">
        <div class="prod-section-header">
          <div>
            <span class="prod-section-title">
              <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"/></svg>
              Variantes y Opciones (Tallas, Colores, etc.)
            </span>
            <p style="margin:2px 0 0 0; font-size:11.5px; color:var(--text-mut);">Configura atributos para generar variantes con precios individuales y varias fotos por variante.</p>
          </div>
        </div>
        
        <div id="variant-attrs-section" class="variant-attrs-box">
          <div id="variant-attrs-container"></div>
          <div id="variant-gen-banner" class="variant-gen-banner"></div>
          <div style="display:flex; justify-content:flex-start; margin-top:0.6rem;">
            <button type="button" class="btn btn-sm btn-primary" id="btn-generate-variants" onclick="handleGenerateVariants()">
              <svg width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" style="margin-right:.25rem"><path d="M12 5v14M5 12h14"/></svg>
              Generar combinaciones
            </button>
          </div>
        </div>
        
        <div id="variants-table-wrapper"></div>
      </div>

    </form>
  `;

  document.body.appendChild(overlay);
}


// ============================================================
//                     SAVE / EDIT / DELETE
// ============================================================
async function saveProduct() {
  if (!FormValidator.validateForm('product-form')) {
    return;
  }

  const variantCards = $$('#variants-list-container .variant-card');
  let totalVariantStock = 0;
  const variantesPayload = variantCards.map(card => {
    let atributos = [];
    try { atributos = JSON.parse(card.dataset.attributes || '[]'); } catch(e) {}
    const stock = Number($('.var-stock', card)?.value) || 0;
    totalVariantStock += stock;
    const idVal = card.dataset.variantId;
    const key = card.dataset.variantKey;

    return {
      id: idVal && idVal !== '' && idVal !== 'null' ? Number(idVal) : null,
      temp_key: key,
      sku: $('.var-sku', card)?.value.trim() || '',
      barcode: $('.var-barcode', card)?.value.trim() || '',
      nombre: card.dataset.variantName || '',
      precio_override: Number($('.var-precio', card)?.value) || 0,
      stock: stock,
      stock_minimo: 5,
      imagen: card.dataset.imageUrls || '',
      activa: $('.var-activo', card)?.checked ?? true,
      atributos: atributos.map(a => ({
        attribute_id: Number(a.attribute_id || a.atributo_id || a.AttributeID || 0),
        attribute_name: String(a.attribute_name || a.atributo_nombre || a.nombre || a.AttributeName || ''),
        value: String(a.value || a.valor || a.Value || '')
      }))
    };
  });

  const body = {
    sku: $('#prod-sku').value.trim(),
    nombre: $('#prod-nombre').value.trim(),
    descripcion: $('#prod-descripcion').value.trim(),
    categoria_id: Number($('#prod-categoria').value) || null,
    estado: $('#prod-estado').value,
    precio: Number($('#prod-precio').value) || 0,
    precio_anterior: Number($('#prod-precio-anterior').value) || null,
    costo: Number($('#prod-costo').value) || null,
    stock: variantCards.length ? totalVariantStock : (Number($('#prod-stock').value) || 0),
    stock_minimo: Number($('#prod-stock-minimo').value) || 5,
    tiene_variantes: variantCards.length > 0,
    variantes: !state.editingProduct ? variantesPayload : undefined
  };

  const btn = $('#btn-save-product');
  btn.disabled = true;
  btn.textContent = 'Guardando…';

  try {
    let res;
    let savedId;

    if (state.editingProduct) {
      savedId = state.editingProduct.id;
      res = await api(`/products/${savedId}`, { method: 'PUT', json: body });

      // Save/Update variants in edit mode
      if (variantesPayload.length > 0) {
        const existing = variantesPayload.filter(v => v.id != null);
        const brandNew = variantesPayload.filter(v => v.id == null);

        if (existing.length > 0) {
          await api(`/products/${savedId}/variants/bulk`, {
            method: 'PUT',
            json: { variants: existing }
          });
          for (const ev of existing) {
            if (ev.id && state.variantPendingFiles[String(ev.id)]?.length) {
              const formData = new FormData();
              state.variantPendingFiles[String(ev.id)].forEach(f => formData.append('images', f));
              await apiUpload(`/products/${savedId}/variants/${ev.id}/images`, formData).catch(() => {});
            }
          }
        }

        for (const nv of brandNew) {
          const created = await api(`/products/${savedId}/variants`, {
            method: 'POST',
            json: {
              sku: nv.sku,
              barcode: nv.barcode,
              nombre: nv.nombre,
              precio_override: nv.precio_override,
              stock: nv.stock,
              stock_minimo: nv.stock_minimo,
              imagen: nv.imagen,
              atributos: nv.atributos
            }
          });
          if (nv.temp_key && state.variantPendingFiles[nv.temp_key]?.length && created.data?.id) {
            const formData = new FormData();
            state.variantPendingFiles[nv.temp_key].forEach(f => formData.append('images', f));
            await apiUpload(`/products/${savedId}/variants/${created.data.id}/images`, formData).catch(() => {});
          }
        }
      }
    } else {
      // Create product + all variants atomically
      res = await api('/products', { method: 'POST', json: body });
      savedId = res.data?.id;

      // Upload pending variant images if any
      if (savedId && res.data?.variantes && res.data.variantes.length) {
        for (let i = 0; i < res.data.variantes.length; i++) {
          const createdVar = res.data.variantes[i];
          const matchingInput = variantesPayload.find(vp => vp.sku === createdVar.sku) || variantesPayload[i];
          if (matchingInput && matchingInput.temp_key && state.variantPendingFiles[matchingInput.temp_key]?.length) {
            const formData = new FormData();
            state.variantPendingFiles[matchingInput.temp_key].forEach(f => formData.append('images', f));
            await apiUpload(`/products/${savedId}/variants/${createdVar.id}/images`, formData).catch(() => {});
          }
        }
      }
    }

    // Upload pending general gallery images if any
    if (state.selectedImages.length && savedId) {
      await uploadImages(savedId);
    }

    Toast.success('Éxito', res.message || 'Producto y variantes guardados correctamente');
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

  if (actions) {
    actions.style.display = 'block';
    const uploadBtn = $('#btn-upload-photos-direct');
    if (uploadBtn) {
      uploadBtn.style.display = state.editingProduct ? 'inline-flex' : 'none';
    }
  }

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
    Toast.info('Aviso', 'Las fotos seleccionadas se guardarán automáticamente al hacer clic en "Guardar"');
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
//                   VARIANT & ATTRIBUTE MANAGEMENT
// ============================================================
async function loadCategoryAttributes(categoryId) {
  if (!categoryId) {
    state.categoryAttributes = [];
    return;
  }

  try {
    const [catRes, globalRes] = await Promise.all([
      api(`/attributes/category/${categoryId}`),
      api('/attributes/global'),
    ]);

    const catAttrs = catRes.data || [];
    const globalAttrs = globalRes.data || [];

    const seen = new Set(catAttrs.map(a => a.id));
    const merged = [...catAttrs];
    globalAttrs.forEach(a => { if (!seen.has(a.id)) merged.push(a); });

    state.categoryAttributes = merged;
    if (merged.length > 0 && state.variantAttributes.length === 0) {
      merged.forEach(ca => {
        const vals = ca.valores || ca.values || [];
        state.variantAttributes.push({
          id: ca.id,
          name: ca.nombre || ca.name,
          values: vals.slice(0, 3),
          suggestions: vals
        });
      });
      renderAttributeBuilder();
    }
  } catch (err) {
    console.error('Error loading attributes:', err);
  }
}

const ATTRIBUTE_PRESETS = [
  { name: 'Talla', defaultVals: ['S', 'M', 'L'], suggestions: ['XS', 'S', 'M', 'L', 'XL', 'XXL', '38', '40', '42'] },
  { name: 'Color', defaultVals: ['Negro', 'Blanco'], suggestions: ['Negro', 'Blanco', 'Azul', 'Rojo', 'Verde', 'Gris', 'Amarillo', 'Rosado', 'Beige'] },
  { name: 'Material', defaultVals: ['Algodón'], suggestions: ['Algodón', 'Poliéster', 'Cuero', 'Sintético', 'Lana', 'Seda'] },
];

function handleAddAttributePreset(presetName) {
  const existing = state.variantAttributes.find(a => a.name.toLowerCase() === presetName.toLowerCase());
  if (existing) {
    Toast.info('Atributo ya añadido', `"${presetName}" ya está en la lista.`);
    return;
  }
  const preset = ATTRIBUTE_PRESETS.find(p => p.name.toLowerCase() === presetName.toLowerCase()) || { name: presetName, defaultVals: [], suggestions: [] };
  state.variantAttributes.push({
    name: preset.name,
    values: [...preset.defaultVals],
    suggestions: [...preset.suggestions]
  });
  renderAttributeBuilder();
}

function handlePromptAddCustomAttribute() {
  const name = prompt('Nombre del nuevo atributo (ej. Capacidad, Estilo, Género):');
  if (!name || !name.trim()) return;
  const cleanName = name.trim();
  const existing = state.variantAttributes.find(a => a.name.toLowerCase() === cleanName.toLowerCase());
  if (existing) {
    Toast.info('Atributo existente', `El atributo "${cleanName}" ya está en la lista.`);
    return;
  }
  state.variantAttributes.push({
    name: cleanName,
    values: [],
    suggestions: []
  });
  renderAttributeBuilder();
}

function handleRemoveAttribute(attrIdx) {
  state.variantAttributes.splice(attrIdx, 1);
  renderAttributeBuilder();
}

function handleAddAttributeValue(attrIdx, val) {
  const attr = state.variantAttributes[attrIdx];
  if (!attr) return;
  if (!attr.values) attr.values = [];
  const cleanVal = String(val).trim();
  if (cleanVal && !attr.values.includes(cleanVal)) {
    attr.values.push(cleanVal);
  }
  renderAttributeBuilder();
}

function handleRemoveAttributeValue(attrIdx, valIdx) {
  const attr = state.variantAttributes[attrIdx];
  if (!attr || !attr.values) return;
  attr.values.splice(valIdx, 1);
  renderAttributeBuilder();
}

function handleAddCustomValueFromInput(attrIdx) {
  const input = $(`#attr-input-${attrIdx}`);
  if (!input) return;
  const val = input.value.trim();
  if (!val) return;
  handleAddAttributeValue(attrIdx, val);
  input.value = '';
}

function renderAttributeBuilder() {
  const container = $('#variant-attrs-container');
  if (!container) return;

  const validAttrs = state.variantAttributes.filter(a => a.values && a.values.length > 0);
  const totalCombos = validAttrs.reduce((acc, a) => acc * a.values.length, validAttrs.length > 0 ? 1 : 0);

  const presetButtons = ATTRIBUTE_PRESETS.map(p => {
    const isAdded = state.variantAttributes.some(a => a.name.toLowerCase() === p.name.toLowerCase());
    return `
      <button type="button" class="btn-attr-preset ${isAdded ? 'btn-attr-preset--active' : ''}" onclick="handleAddAttributePreset('${p.name}')">
        <svg width="12" height="12" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
        ${p.name}
      </button>
    `;
  }).join('');

  let html = `
    <div class="attr-builder-top">
      <span class="attr-builder-label">Atributos disponibles:</span>
      <div class="attr-presets-row">
        ${presetButtons}
        <button type="button" class="btn-attr-preset" onclick="handlePromptAddCustomAttribute()">
          <svg width="12" height="12" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          Otro
        </button>
      </div>
    </div>
  `;

  if (state.variantAttributes.length === 0) {
    html += `
      <div class="attr-empty-prompt">
        <svg width="18" height="18" fill="none" stroke="var(--accent)" stroke-width="2" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
        <span>Haz clic en <strong>+ Talla</strong> o <strong>+ Color</strong> arriba para definir las opciones de tu producto.</span>
      </div>
    `;
  } else {
    html += '<div class="attr-rows-list">';
    state.variantAttributes.forEach((attr, attrIdx) => {
      const matchedPreset = ATTRIBUTE_PRESETS.find(p => p.name.toLowerCase() === attr.name.toLowerCase());
      const suggestions = attr.suggestions || matchedPreset?.suggestions || [];
      const unused = suggestions.filter(s => !(attr.values || []).includes(s));

      const tags = (attr.values || []).map((val, valIdx) => `
        <span class="attr-val-tag">
          ${esc(val)}
          <button type="button" class="attr-val-tag-del" onclick="handleRemoveAttributeValue(${attrIdx}, ${valIdx})" title="Quitar">&times;</button>
        </span>
      `).join('');

      const suggestionPills = unused.slice(0, 7).map(s => `
        <button type="button" class="attr-suggestion-chip" onclick="handleAddAttributeValue(${attrIdx}, '${esc(s)}')">+ ${esc(s)}</button>
      `).join('');

      html += `
        <div class="attr-row-card">
          <div class="attr-row-header">
            <div class="attr-row-title">
              <span class="attr-icon-dot"></span>
              <strong>${esc(attr.name)}</strong>
              <span class="attr-row-badge">${(attr.values || []).length} opciones</span>
            </div>
            <button type="button" class="btn-attr-row-del" onclick="handleRemoveAttribute(${attrIdx})" title="Eliminar atributo">
              <svg width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
            </button>
          </div>

          <div class="attr-row-body">
            <div class="attr-val-tags-container">
              ${tags || '<span class="attr-val-empty-hint">Sin opciones agregadas. Escribe o selecciona sugerencias abajo:</span>'}
            </div>

            <div class="attr-input-group">
              <input type="text" id="attr-input-${attrIdx}" class="form-input form-input--sm attr-custom-input" placeholder="Escribe una opción (ej. ${attr.name === 'Talla' ? 'XL, 38' : (attr.name === 'Color' ? 'Azul Marino' : 'Opción')}) y presiona Enter" onkeydown="if(event.key==='Enter'){event.preventDefault();handleAddCustomValueFromInput(${attrIdx});}">
              <button type="button" class="btn btn-sm btn-secondary" onclick="handleAddCustomValueFromInput(${attrIdx})">Agregar</button>
            </div>

            ${suggestionPills ? `
              <div class="attr-suggestions-bar">
                <span class="attr-suggestions-label">Sugerencias:</span>
                <div class="attr-suggestions-list">${suggestionPills}</div>
              </div>
            ` : ''}
          </div>
        </div>
      `;
    });
    html += '</div>';
  }

  container.innerHTML = html;

  // Banner status
  const banner = $('#variant-gen-banner');
  if (banner) {
    if (totalCombos > 0) {
      const summaryParts = validAttrs.map(a => `${a.values.length} ${a.name.toLowerCase()}`);
      banner.className = 'variant-gen-banner variant-gen-banner--ready';
      banner.innerHTML = `
        <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
        <span>Se generarán <strong>${totalCombos} combinaciones</strong> (${summaryParts.join(' × ')}). Cada variante tendrá precio, stock y fotos individuales.</span>
      `;
    } else {
      banner.className = 'variant-gen-banner variant-gen-banner--empty';
      banner.innerHTML = `
        <svg width="15" height="15" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
        <span>Agrega al menos una opción a tus atributos para generar las variantes disponibles.</span>
      `;
    }
  }

  // Button text
  const genBtn = $('#btn-generate-variants');
  if (genBtn) {
    genBtn.innerHTML = `
      <svg width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" style="margin-right:.25rem"><path d="M12 5v14M5 12h14"/></svg>
      Generar ${totalCombos > 0 ? `${totalCombos} ` : ''}combinaciones
    `;
  }
}

function handleGenerateVariants() {
  const validAttrs = state.variantAttributes.filter(a => a.values && a.values.length > 0);
  const totalCombos = validAttrs.reduce((acc, a) => acc * a.values.length, validAttrs.length > 0 ? 1 : 0);

  if (totalCombos === 0) {
    Toast.error('Atributos requeridos', 'Añade al menos un atributo (ej. Talla o Color) con sus opciones antes de generar combinaciones.');
    return;
  }

  const baseSku = $('#prod-sku')?.value.trim() || 'PROD';
  const baseName = $('#prod-nombre')?.value.trim() || '';
  const basePrice = Number($('#prod-precio')?.value) || 0;
  const baseStockMinimo = Number($('#prod-stock-minimo')?.value) || 5;

  let combinations = [[]];
  for (const attr of validAttrs) {
    const nextCombos = [];
    for (const combo of combinations) {
      for (const val of attr.values) {
        nextCombos.push([
          ...combo,
          {
            attribute_id: attr.id || 0,
            attribute_name: attr.name,
            value: val
          }
        ]);
      }
    }
    combinations = nextCombos;
  }

  syncVariantsFromDOM();

  const generated = combinations.map((combo, idx) => {
    const skuSuffix = combo.map(a => a.value.replace(/[^a-zA-Z0-9-]/g, '').toUpperCase()).join('-');
    const cleanBaseSku = baseSku.replace(/[^a-zA-Z0-9-]/g, '').toUpperCase();
    const varSku = cleanBaseSku ? `${cleanBaseSku}-${skuSuffix}` : skuSuffix;

    const nameSuffix = combo.map(a => `${a.attribute_name}: ${a.value}`).join(', ');
    const varName = baseName ? `${baseName} (${nameSuffix})` : nameSuffix;

    const existing = state.existingVariants.find(ev => ev.sku === varSku);
    if (existing) {
      return {
        ...existing,
        nombre: varName,
        atributos: combo
      };
    }

    return {
      id: null,
      temp_key: 'tmp_' + Date.now() + '_' + idx,
      sku: varSku,
      barcode: '',
      nombre: varName,
      precio: basePrice,
      precio_override: basePrice,
      stock: 0,
      stock_minimo: baseStockMinimo,
      imagen: '',
      activa: true,
      atributos: combo
    };
  });

  state.existingVariants = generated;
  renderVariantsTable(state.existingVariants, basePrice);
  Toast.success('Variantes generadas', `${generated.length} combinaciones listas. Ahora puedes ajustar los precios individuales, stock y subir fotos por variante.`);
}

async function loadVariants(productId) {
  const wrapper = $('#variants-table-wrapper');
  if (!wrapper) return;
  setLoading(wrapper, true);

  try {
    const res = await api(`/products/${productId}/variants`);
    state.existingVariants = res.data || [];
    const basePrice = state.editingProduct?.precio || 0;

    // Reconstruct attributes from existing variants
    const attrMap = {};
    state.existingVariants.forEach(v => {
      (v.atributos || []).forEach(a => {
        const name = a.attribute_name || a.atributo_nombre || a.nombre || a.AttributeName || 'Atributo';
        const val = a.value || a.valor || a.Value || '';
        if (!attrMap[name]) attrMap[name] = new Set();
        if (val) attrMap[name].add(val);
      });
    });

    state.variantAttributes = Object.entries(attrMap).map(([name, valSet]) => ({
      name: name,
      values: Array.from(valSet),
      suggestions: []
    }));

    renderAttributeBuilder();
    renderVariantsTable(state.existingVariants, basePrice);
  } catch (err) {
    wrapper.innerHTML = emptyState('Error al cargar variantes', 'warning');
  }
}

function renderVariantsTable(variants, basePrice) {
  const wrapper = $('#variants-table-wrapper');
  if (!wrapper) return;

  if (!variants.length) {
    wrapper.innerHTML = emptyState('Sin variantes generadas aún', 'variantEmpty');
    $('#prod-stock').disabled = false;
    return;
  }

  $('#prod-stock').disabled = true;

  let totalStock = 0;
  const cardsHTML = variants.map((v, i) => {
    const key = v.id ? String(v.id) : (v.temp_key || ('tmp_' + i));
    const attrChips = (v.atributos || []).map(a => {
      const attrName = a.attribute_name || a.atributo_nombre || a.nombre || a.AttributeName || '';
      const attrVal = a.value || a.valor || a.Value || '';
      const color = chipColor(attrName);
      return `<span class="variant-chip" style="background:${color}">${esc(attrName)}: ${esc(attrVal)}</span>`;
    }).join('');

    const stock = Number(v.stock) || 0;
    totalStock += stock;
    const active = v.activa !== false && v.activo !== false;

    // Saved images on server
    const savedImages = v.imagen ? v.imagen.split(';').filter(x => x) : [];
    // Pending local files
    const pendingFiles = state.variantPendingFiles[key] || [];
    const totalImgs = savedImages.length + pendingFiles.length;

    let imgThumbnails = '';
    savedImages.forEach((url, imgIdx) => {
      imgThumbnails += `
        <div class="var-img-thumbnail">
          <img src="${url}" alt="">
          <button type="button" class="var-img-remove-btn" onclick="handleRemoveVariantImage('${key}', ${imgIdx}, false)" title="Quitar foto">&times;</button>
        </div>
      `;
    });
    pendingFiles.forEach((file, fIdx) => {
      const localUrl = URL.createObjectURL(file);
      imgThumbnails += `
        <div class="var-img-thumbnail var-img-thumbnail--pending" title="${esc(file.name)}">
          <img src="${localUrl}" alt="">
          <button type="button" class="var-img-remove-btn" onclick="handleRemoveVariantImage('${key}', ${fIdx}, true)" title="Quitar foto">&times;</button>
        </div>
      `;
    });

    return `
      <div class="variant-card" data-variant-id="${v.id || ''}" data-variant-key="${key}" data-index="${i}" data-variant-name="${esc(v.nombre || '')}" data-image-urls="${esc(v.imagen || '')}" data-attributes='${esc(JSON.stringify(v.atributos || []))}'>
        
        <!-- Header -->
        <div class="variant-card__header">
          <div class="variant-card__chips">
            ${attrChips || '<span class="variant-chip variant-chip--default">General</span>'}
          </div>
          <div class="variant-card__actions">
            <label class="toggle-switch" title="Activa / Inactiva">
              <input type="checkbox" class="var-activo" ${active ? 'checked' : ''}>
              <span class="toggle-slider"></span>
            </label>
            <span class="variant-card__status-text">${active ? 'Activa' : 'Inactiva'}</span>
            <button type="button" class="btn-var-delete" onclick="handleDeleteVariant('${key}')" title="Eliminar variante">
              <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
            </button>
          </div>
        </div>

        <!-- Body -->
        <div class="variant-card__body">
          <!-- Multi-Images per Variant -->
          <div class="variant-card__imgs">
            <div style="display:flex; justify-content:space-between; align-items:center;">
              <label class="form-label form-label--xs" style="margin:0;">Fotos (${totalImgs})</label>
            </div>
            <div class="variant-card__img-list">
              ${imgThumbnails}
              <button type="button" class="var-img-add-btn" onclick="$('#var-file-input-${key}').click()" title="Subir varias fotos a esta variante">
                <svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                <span>+ Foto</span>
              </button>
              <input type="file" id="var-file-input-${key}" accept="image/*" multiple style="display:none;" onchange="handleUploadVariantImages(this, '${key}')">
            </div>
          </div>

          <!-- Fields -->
          <div class="variant-card__fields">
            <div class="form-group" style="margin:0;">
              <label class="form-label form-label--xs">SKU *</label>
              <input type="text" class="form-input form-input--sm var-sku" value="${esc(v.sku || '')}" placeholder="SKU" required data-validate="required">
            </div>
            <div class="form-group" style="margin:0;">
              <label class="form-label form-label--xs">Código barras</label>
              <input type="text" class="form-input form-input--sm var-barcode" value="${esc(v.barcode || '')}" placeholder="EAN / UPC">
            </div>
            <div class="form-group" style="margin:0;">
              <label class="form-label form-label--xs">Precio individual ($) *</label>
              <input type="number" class="form-input form-input--sm var-precio" value="${v.precio_override || v.precio || basePrice || ''}" min="0" step="1" placeholder="0" required data-validate="required|number">
            </div>
            <div class="form-group" style="margin:0;">
              <label class="form-label form-label--xs">Stock *</label>
              <input type="number" class="form-input form-input--sm var-stock" value="${stock}" min="0" step="1" placeholder="0" oninput="updateTotalStockFromVariants()" required data-validate="required|number">
            </div>
          </div>
        </div>

      </div>
    `;
  }).join('');

  wrapper.innerHTML = `
    <div style="margin-top:0.75rem;">
      <div style="margin-bottom:0.5rem; display:flex; justify-content:space-between; align-items:center;">
        <span style="font-size:12px; font-weight:700; color:var(--text);">Variantes configuradas (${variants.length})</span>
        <span style="font-size:11px; color:var(--text-sec); background:var(--bg-surf); padding:3px 10px; border-radius:12px; border:1px solid var(--border);">Stock Total Sumado: <strong id="var-total-stock-badge" style="color:var(--accent)">${totalStock}</strong></span>
      </div>
      <div id="variants-list-container">
        ${cardsHTML}
      </div>
    </div>
  `;

  $('#prod-stock').value = totalStock;
}

function updateTotalStockFromVariants() {
  const inputs = $$('#variants-list-container .var-stock');
  let total = 0;
  inputs.forEach(inp => { total += Number(inp.value) || 0; });
  const stockField = $('#prod-stock');
  if (stockField) stockField.value = total;
  const badge = $('#var-total-stock-badge');
  if (badge) badge.textContent = total;
}

function syncVariantsFromDOM() {
  const cards = $$('#variants-list-container .variant-card');
  if (!cards.length) return;
  state.existingVariants = cards.map(card => {
    let atributos = [];
    try { atributos = JSON.parse(card.dataset.attributes || '[]'); } catch(e) {}
    const idVal = card.dataset.variantId;
    return {
      id: idVal && idVal !== '' && idVal !== 'null' ? Number(idVal) : null,
      temp_key: card.dataset.variantKey || '',
      sku: $('.var-sku', card)?.value.trim() || '',
      barcode: $('.var-barcode', card)?.value.trim() || '',
      nombre: card.dataset.variantName || '',
      precio_override: Number($('.var-precio', card)?.value) || 0,
      stock: Number($('.var-stock', card)?.value) || 0,
      imagen: card.dataset.imageUrls || '',
      activa: $('.var-activo', card)?.checked ?? true,
      atributos: atributos
    };
  });
}

/** Generate a pastel background color deterministically from a string */
function chipColor(str) {
  let hash = 0;
  for (let i = 0; i < str.length; i++) hash = str.charCodeAt(i) + ((hash << 5) - hash);
  const h = Math.abs(hash) % 360;
  return `hsl(${h}, 55%, 25%)`;
}

async function handleUploadVariantImages(fileInput, variantKey) {
  if (!fileInput.files.length) return;
  const isSaved = state.editingProduct && !variantKey.startsWith('tmp_');

  if (isSaved) {
    const formData = new FormData();
    for (const file of fileInput.files) {
      formData.append('images', file);
    }
    fileInput.disabled = true;

    try {
      await apiUpload(`/products/${state.editingProduct.id}/variants/${variantKey}/images`, formData);
      Toast.success('Éxito', 'Imágenes subidas para la variante');
      loadVariants(state.editingProduct.id);
    } catch (err) {
      Toast.error('Error', err.message);
    } finally {
      fileInput.disabled = false;
      fileInput.value = '';
    }
  } else {
    // Store in memory (multiple files allowed)
    if (!state.variantPendingFiles[variantKey]) state.variantPendingFiles[variantKey] = [];
    for (const file of fileInput.files) {
      if (file.size > 5 * 1024 * 1024) {
        Toast.error('Archivo grande', `"${file.name}" supera el límite de 5 MB`);
        continue;
      }
      state.variantPendingFiles[variantKey].push(file);
    }
    fileInput.value = '';
    syncVariantsFromDOM();
    renderVariantsTable(state.existingVariants, Number($('#prod-precio')?.value) || 0);
  }
}

async function handleRemoveVariantImage(variantKey, imageIndex, isPending) {
  if (isPending) {
    if (state.variantPendingFiles[variantKey]) {
      state.variantPendingFiles[variantKey].splice(imageIndex, 1);
    }
    syncVariantsFromDOM();
    renderVariantsTable(state.existingVariants, Number($('#prod-precio')?.value) || 0);
    return;
  }

  if (!state.editingProduct) return;
  const variantId = Number(variantKey);
  const variant = state.existingVariants.find(v => v.id === variantId);
  if (!variant) return;

  const confirmed = await window.confirmDelete({
    title: 'Eliminar imagen',
    message: '¿Estás seguro de que deseas eliminar esta imagen de la variante?',
  });
  if (!confirmed) return;

  const images = variant.imagen ? variant.imagen.split(';').filter(x => x) : [];
  images.splice(imageIndex, 1);
  const newImageString = images.join(';');

  try {
    await api(`/products/${state.editingProduct.id}/variants/${variantId}`, {
      method: 'PUT',
      json: { imagen: newImageString }
    });
    Toast.success('Éxito', 'Imagen eliminada de la variante');
    loadVariants(state.editingProduct.id);
  } catch (err) {
    Toast.error('Error', err.message);
  }
}

async function handleDeleteVariant(variantKey) {
  const isSaved = state.editingProduct && !variantKey.startsWith('tmp_');

  if (isSaved) {
    const confirmed = await window.confirmDelete({
      title: 'Eliminar variante',
      message: '¿Eliminar esta variante permanentemente?',
    });
    if (!confirmed) return;

    try {
      await api(`/products/${state.editingProduct.id}/variants/${variantKey}`, { method: 'DELETE' });
      Toast.success('Variante', 'Variante eliminada');
      loadVariants(state.editingProduct.id);
    } catch (err) {
      Toast.error('Error', err.message);
    }
  } else {
    // Remove in memory
    syncVariantsFromDOM();
    state.existingVariants = state.existingVariants.filter((v, idx) => {
      const key = v.id ? String(v.id) : (v.temp_key || ('tmp_' + idx));
      return key !== variantKey;
    });
    delete state.variantPendingFiles[variantKey];
    renderVariantsTable(state.existingVariants, Number($('#prod-precio')?.value) || 0);
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
      width: 40px; height: 40px; border-radius: 6px; object-fit: cover;
      border: 1px solid var(--border); flex-shrink: 0;
    }
    .prod-thumb--placeholder {
      display: flex; align-items: center; justify-content: center;
      background: var(--bg-surf); color: var(--text-mut);
    }

    /* ---- Modal sizing & layout ---- */
    .modal--product {
      max-width: 960px;
      width: 95vw;
      max-height: 92vh;
      display: flex;
      flex-direction: column;
    }
    .modal--product .modal-header {
      padding: 12px 18px;
      flex-shrink: 0;
    }
    .modal--product .modal-footer {
      padding: 12px 18px;
      flex-shrink: 0;
    }
    .modal--product .modal-body {
      padding: 14px 18px;
      max-height: calc(92vh - 115px);
      overflow-y: auto;
    }

    /* ---- Form Layout ---- */
    .prod-modal-form {
      display: flex;
      flex-direction: column;
      gap: 1.1rem;
    }

    .prod-form-section {
      display: flex;
      flex-direction: column;
      gap: 0.75rem;
      padding-bottom: 1rem;
      border-bottom: 1px solid var(--border);
    }
    .prod-form-section:last-child {
      border-bottom: none;
      padding-bottom: 0;
    }

    .prod-section-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 0.5rem;
    }
    .prod-section-title {
      font-size: 0.88rem;
      font-weight: 700;
      color: var(--text);
      display: flex;
      align-items: center;
      gap: 0.45rem;
      letter-spacing: -0.01em;
    }
    .prod-section-title svg {
      color: var(--accent);
      flex-shrink: 0;
    }

    /* ---- Grids ---- */
    .prod-grid {
      display: grid;
      gap: 0.65rem;
    }
    .prod-grid .form-group {
      margin-bottom: 0;
    }

    .prod-grid--header {
      grid-template-columns: 130px 1.5fr 1.1fr 110px;
    }

    .prod-grid--pricing {
      grid-template-columns: repeat(5, 1fr);
    }

    .prod-textarea {
      min-height: 52px;
      resize: vertical;
    }

    /* ---- Image Layout ---- */
    .prod-img-layout {
      display: grid;
      grid-template-columns: 210px 1fr;
      gap: 0.85rem;
      align-items: start;
    }
    .img-upload-zone {
      border: 1.5px dashed var(--border);
      border-radius: 8px;
      padding: 0.75rem 0.5rem;
      text-align: center;
      cursor: pointer;
      transition: all .2s;
      background: var(--bg-surf);
      min-height: 80px;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    .img-upload-zone:hover,
    .img-upload-zone.dragover {
      border-color: var(--accent);
      background: var(--accent-dim);
    }
    .img-upload-zone__text {
      margin: 0.25rem 0 0;
      color: var(--text-sec);
      font-size: 11.5px;
      line-height: 1.3;
    }
    .img-upload-link {
      color: var(--accent);
      font-weight: 600;
      text-decoration: underline;
    }
    .img-upload-zone__hint {
      font-size: 0.65rem;
      color: var(--text-mut);
      display: block;
      margin-top: 2px;
    }

    .prod-img-content {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
    }
    .img-preview-grid,
    .img-existing-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(72px, 1fr));
      gap: 0.5rem;
    }

    .img-preview-item,
    .img-existing-item {
      position: relative;
      border-radius: 6px;
      overflow: hidden;
      aspect-ratio: 1;
      border: 1px solid var(--border);
      background: var(--bg-surf);
    }
    .img-existing-item--principal {
      border-color: var(--warning);
      box-shadow: 0 0 0 1px var(--warning);
    }
    .img-preview-item img,
    .img-existing-item img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
    .img-preview-remove {
      position: absolute;
      top: 2px;
      right: 2px;
      width: 18px;
      height: 18px;
      border-radius: 50%;
      background: rgba(0,0,0,.75);
      color: #fff;
      border: none;
      cursor: pointer;
      font-size: 12px;
      line-height: 1;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    .img-preview-name {
      position: absolute;
      bottom: 0;
      left: 0;
      right: 0;
      padding: 1px 3px;
      font-size: 0.6rem;
      color: #fff;
      background: rgba(0,0,0,.65);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .img-star-badge {
      position: absolute;
      top: 3px;
      left: 3px;
      font-size: 0.75rem;
      filter: drop-shadow(0 1px 2px rgba(0,0,0,.7));
    }
    .img-hover-overlay {
      position: absolute;
      inset: 0;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 0.35rem;
      background: rgba(0,0,0,.6);
      opacity: 0;
      transition: opacity .15s;
    }
    .img-existing-item:hover .img-hover-overlay { opacity: 1; }
    @media (hover: none) {
      .img-hover-overlay { opacity: 1; background: rgba(0,0,0,.35); }
    }
    .img-action-btn {
      width: 28px;
      height: 28px;
      border-radius: 50%;
      border: none;
      background: rgba(255,255,255,.2);
      color: #fff;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;
      backdrop-filter: blur(4px);
      transition: background .15s;
    }
    .img-action-btn:hover { background: rgba(255,255,255,.4); }
    .img-action-btn--danger:hover { background: rgba(220,38,38,.8); }

    /* ---- Variant & Attribute Section ---- */
    .variant-attrs-box {
      background: var(--bg-surf);
      border: 1px solid var(--border);
      border-radius: 8px;
      padding: 0.75rem 0.95rem;
      display: flex;
      flex-direction: column;
      gap: 0.65rem;
    }
    .attr-builder-top {
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      gap: 0.5rem;
    }
    .attr-builder-label {
      font-size: 11.5px;
      font-weight: 600;
      color: var(--text-sec);
    }
    .attr-presets-row {
      display: flex;
      flex-wrap: wrap;
      gap: 0.35rem;
    }
    .btn-attr-preset {
      display: inline-flex;
      align-items: center;
      gap: 0.25rem;
      padding: 3px 9px;
      font-size: 11.5px;
      font-weight: 500;
      border-radius: 14px;
      border: 1px dashed var(--border);
      background: var(--bg-elev);
      color: var(--text-sec);
      cursor: pointer;
      transition: all .15s ease;
    }
    .btn-attr-preset:hover {
      border-color: var(--accent);
      color: var(--accent);
      background: var(--accent-dim);
    }
    .btn-attr-preset--active {
      border-style: solid;
      border-color: var(--accent);
      color: var(--accent);
      background: var(--accent-dim);
      font-weight: 600;
    }

    .attr-empty-prompt {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      padding: 0.6rem 0.8rem;
      background: var(--bg-elev);
      border: 1px dashed var(--border);
      border-radius: 6px;
      font-size: 12px;
      color: var(--text-sec);
    }

    .attr-rows-list {
      display: flex;
      flex-direction: column;
      gap: 0.55rem;
    }
    .attr-row-card {
      background: var(--bg-elev);
      border: 1px solid var(--border);
      border-radius: 6px;
      padding: 0.55rem 0.75rem;
      display: flex;
      flex-direction: column;
      gap: 0.45rem;
    }
    .attr-row-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
    }
    .attr-row-title {
      display: flex;
      align-items: center;
      gap: 0.4rem;
      font-size: 12px;
      color: var(--text);
    }
    .attr-icon-dot {
      width: 7px;
      height: 7px;
      border-radius: 50%;
      background: var(--accent);
      display: inline-block;
    }
    .attr-row-badge {
      font-size: 10px;
      padding: 1px 6px;
      border-radius: 10px;
      background: var(--bg-surf);
      color: var(--text-mut);
      border: 1px solid var(--border);
    }
    .btn-attr-row-del {
      background: none;
      border: none;
      color: var(--danger);
      cursor: pointer;
      padding: 2px;
      border-radius: 4px;
      opacity: 0.75;
      transition: opacity .15s;
    }
    .btn-attr-row-del:hover { opacity: 1; }

    .attr-row-body {
      display: flex;
      flex-direction: column;
      gap: 0.4rem;
    }
    .attr-val-tags-container {
      display: flex;
      flex-wrap: wrap;
      gap: 0.3rem;
      align-items: center;
      min-height: 24px;
    }
    .attr-val-empty-hint {
      font-size: 11px;
      color: var(--text-mut);
      font-style: italic;
    }
    .attr-val-tag {
      display: inline-flex;
      align-items: center;
      gap: 0.25rem;
      padding: 2px 7px;
      background: var(--accent);
      color: #fff;
      border-radius: 12px;
      font-size: 11px;
      font-weight: 600;
    }
    .attr-val-tag-del {
      background: none;
      border: none;
      color: #fff;
      cursor: pointer;
      font-size: 13px;
      line-height: 1;
      padding: 0;
      display: flex;
      align-items: center;
      opacity: 0.8;
    }
    .attr-val-tag-del:hover { opacity: 1; }

    .attr-input-group {
      display: flex;
      gap: 0.35rem;
    }
    .attr-custom-input {
      max-width: 320px;
      flex: 1;
    }

    .attr-suggestions-bar {
      display: flex;
      align-items: center;
      gap: 0.4rem;
      flex-wrap: wrap;
    }
    .attr-suggestions-label {
      font-size: 10.5px;
      color: var(--text-mut);
      font-weight: 500;
    }
    .attr-suggestions-list {
      display: flex;
      flex-wrap: wrap;
      gap: 0.25rem;
    }
    .attr-suggestion-chip {
      padding: 1px 6px;
      border-radius: 10px;
      font-size: 10.5px;
      background: var(--bg-surf);
      border: 1px solid var(--border);
      color: var(--text-sec);
      cursor: pointer;
      transition: all .15s;
    }
    .attr-suggestion-chip:hover {
      border-color: var(--accent);
      color: var(--accent);
      background: var(--accent-dim);
    }

    /* ---- Variant Generation Banner ---- */
    .variant-gen-banner {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      padding: 0.5rem 0.75rem;
      border-radius: 6px;
      font-size: 11.5px;
      line-height: 1.4;
    }
    .variant-gen-banner--ready {
      background: rgba(16, 185, 129, 0.1);
      border: 1px solid rgba(16, 185, 129, 0.3);
      color: #10b981;
    }
    .variant-gen-banner--empty {
      background: var(--bg-elev);
      border: 1px solid var(--border);
      color: var(--text-sec);
    }

    /* ---- Variant Card ---- */
    .variant-card {
      background: var(--bg-surf);
      border: 1px solid var(--border);
      border-radius: 8px;
      padding: 0.65rem 0.85rem;
      display: flex;
      flex-direction: column;
      gap: 0.55rem;
      margin-bottom: 0.55rem;
      transition: border-color .2s, box-shadow .2s;
    }
    .variant-card:hover {
      border-color: var(--accent);
    }
    .variant-card__header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding-bottom: 0.4rem;
      border-bottom: 1px solid var(--border);
    }
    .variant-card__chips {
      display: flex;
      flex-wrap: wrap;
      gap: 0.25rem;
      align-items: center;
    }
    .variant-chip {
      display: inline-block;
      padding: 1px 7px;
      border-radius: 10px;
      font-size: 10.5px;
      font-weight: 600;
      color: #fff;
    }
    .variant-chip--default {
      background: #4b5563;
    }
    .variant-card__actions {
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }
    .variant-card__status-text {
      font-size: 11.5px;
      font-weight: 500;
      color: var(--text-sec);
    }
    .btn-var-delete {
      background: none;
      border: none;
      color: var(--danger);
      cursor: pointer;
      padding: 3px;
      border-radius: 4px;
      display: flex;
      align-items: center;
      justify-content: center;
      opacity: 0.8;
      transition: opacity .15s;
    }
    .btn-var-delete:hover { opacity: 1; }

    .variant-card__body {
      display: grid;
      grid-template-columns: auto 1fr;
      gap: 0.85rem;
      align-items: center;
    }
    .variant-card__imgs {
      display: flex;
      flex-direction: column;
      gap: 0.25rem;
    }
    .variant-card__img-list {
      display: flex;
      flex-wrap: wrap;
      gap: 0.35rem;
      align-items: center;
    }
    .var-img-thumbnail {
      position: relative;
      width: 42px;
      height: 42px;
      border-radius: 6px;
      overflow: hidden;
      border: 1px solid var(--border);
      background: var(--bg-elev);
      flex-shrink: 0;
    }
    .var-img-thumbnail img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
    .var-img-remove-btn {
      position: absolute;
      top: 1px;
      right: 1px;
      background: rgba(0,0,0,0.8);
      color: #fff;
      border: none;
      width: 14px;
      height: 14px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 9px;
      cursor: pointer;
      line-height: 1;
    }
    .var-img-remove-btn:hover {
      background: var(--danger);
    }
    .var-img-add-btn {
      width: 42px;
      height: 42px;
      border-radius: 6px;
      border: 1.5px dashed var(--border);
      background: var(--bg-elev);
      display: flex;
      align-items: center;
      justify-content: center;
      color: var(--text-mut);
      cursor: pointer;
      transition: all .2s;
      flex-shrink: 0;
    }
    .var-img-add-btn:hover {
      border-color: var(--accent);
      color: var(--accent);
    }

    .variant-card__fields {
      display: grid;
      grid-template-columns: 1.2fr 1.2fr 1fr 0.9fr;
      gap: 0.5rem;
    }
    .form-label--xs {
      font-size: 10.5px;
      margin-bottom: 2px;
      font-weight: 600;
      color: var(--text-sec);
    }
    .form-input--sm {
      height: 32px;
      padding: 4px 8px;
      font-size: 12.5px;
    }

    /* ---- Toggle switch ---- */
    .toggle-switch {
      position: relative;
      display: inline-block;
      width: 32px;
      height: 18px;
    }
    .toggle-switch input { opacity: 0; width: 0; height: 0; }
    .toggle-slider {
      position: absolute; cursor: pointer; inset: 0;
      background: var(--border); border-radius: 18px; transition: .2s;
    }
    .toggle-slider::before {
      content: ''; position: absolute; width: 12px; height: 12px;
      left: 3px; bottom: 3px; background: #fff; border-radius: 50%;
      transition: .2s;
    }
    .toggle-switch input:checked + .toggle-slider { background: var(--accent); }
    .toggle-switch input:checked + .toggle-slider::before { transform: translateX(14px); }

    /* ---- Spin animation ---- */
    @keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

    /* ---- Responsive Breakpoints ---- */
    @media (max-width: 860px) {
      .prod-grid--header {
        grid-template-columns: 1fr 1fr;
      }
      .prod-grid--pricing {
        grid-template-columns: repeat(3, 1fr);
      }
      .prod-img-layout {
        grid-template-columns: 1fr;
      }
      .variant-card__body {
        grid-template-columns: 1fr;
        gap: 0.5rem;
      }
    }

    @media (max-width: 580px) {
      .modal--product {
        width: 98vw;
        max-height: 95vh;
        margin: 4px;
      }
      .modal--product .modal-header,
      .modal--product .modal-footer {
        padding: 10px 14px;
      }
      .modal--product .modal-body {
        padding: 10px 12px;
        max-height: calc(95vh - 105px);
      }
      .prod-grid--header {
        grid-template-columns: 1fr;
      }
      .prod-grid--pricing {
        grid-template-columns: 1fr 1fr;
      }
      .variant-card__fields {
        grid-template-columns: 1fr 1fr;
      }
    }
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
