let posProducts = [];
let posCategories = [];
let cart = [];
let selectedCustomer = null;
let selectedCategory = 0;
let selectedPayMethod = 'efectivo';

async function loadPOSProducts() {
    try {
        const r = await fetch(`${API_URL}/pos/products?limit=100&estado=activo`, {
            headers: { 'Authorization': 'Bearer ' + token }
        });
        if (r.ok) {
            const res = await r.json();
            posProducts = res.data || [];
            renderPOSProducts();
        }
    } catch (e) {
        console.error("Error cargando productos", e);
    }
}

async function loadPOSCategories() {
    try {
        const r = await fetch(`${API_URL}/categories`, {
            headers: { 'Authorization': 'Bearer ' + token }
        });
        if (r.ok) {
            const res = await r.json();
            posCategories = res.data || [];
            renderPOSCategories();
        }
    } catch (e) {
        console.error("Error cargando categorías", e);
    }
}

function renderPOSCategories() {
    const container = document.getElementById('pos-categories');
    if (!container) return;

    const allClass = selectedCategory === 0 ? 'btn-primary' : 'btn-secondary';
    let html = `<button class="btn btn-sm ${allClass}" onclick="selectCategory(0)" style="border-radius:20px; font-size:12px; height:32px; white-space:nowrap; flex-shrink:0;">Todos</button>`;

    html += posCategories.map(c => {
        const btnClass = selectedCategory === c.id ? 'btn-primary' : 'btn-secondary';
        return `<button class="btn btn-sm ${btnClass}" onclick="selectCategory(${c.id})" style="border-radius:20px; font-size:12px; height:32px; white-space:nowrap; flex-shrink:0;">${c.nombre}</button>`;
    }).join('');

    container.innerHTML = html;
}

function selectCategory(id) {
    selectedCategory = id;
    renderPOSCategories();
    renderPOSProducts();
}

function renderPOSProducts() {
    const grid = document.getElementById('pos-products-grid');
    if (!grid) return;

    const q = document.getElementById('pos-search').value.toLowerCase();
    let filtered = posProducts;

    // Filtrar por término de búsqueda (nombre o SKU)
    if (q) {
        filtered = filtered.filter(p =>
            p.nombre.toLowerCase().includes(q) ||
            (p.sku && p.sku.toLowerCase().includes(q))
        );
    }

    // Filtrar por categoría seleccionada
    if (selectedCategory > 0) {
        filtered = filtered.filter(p => p.categoria_id === selectedCategory);
    }

    if (!filtered.length) {
        grid.innerHTML = '<div style="text-align:center;padding:40px;color:var(--text-mut);grid-column:1/-1;">Sin productos en esta sección</div>';
        return;
    }

    grid.innerHTML = filtered.map(p => {
        const isOutOfStock = p.stock <= 0 && !p.permite_stock_negativo;
        const cardStyle = isOutOfStock ? 'opacity: 0.55; cursor: not-allowed;' : 'cursor: pointer;';
        const clickAction = isOutOfStock ? '' : `onclick="addToCart(${p.id})"`;
        const badgeClass = p.stock <= 0 ? 'badge-red' : (p.stock <= p.stock_minimo ? 'badge-yellow' : 'badge-green');
        const badgeText = p.stock <= 0 ? 'Agotado' : `Stock: ${p.stock}`;

        const variantsIndicator = p.tiene_variantes ? `<div style="font-size:11px;color:var(--accent);font-weight:600;margin-top:2px;">Múltiples Variantes</div>` : '';

        return `
            <div class="card" style="padding:14px; display:flex; flex-direction:column; justify-content:space-between; height:100%; transition:transform 0.15s, box-shadow 0.15s; border:1px solid var(--border); ${cardStyle}" ${clickAction} onmouseover="this.style.transform='translateY(-2px)';" onmouseout="this.style.transform='none';">
                <div>
                    <div style="font-weight:600; font-size:13px; color:var(--text); line-height:1.4; display:-webkit-box; -webkit-line-clamp:2; -webkit-box-orient:vertical; overflow:hidden; min-height:36px; margin-bottom:6px;">${p.nombre}</div>
                    <div class="mono" style="font-size:14px; font-weight:700; color:var(--accent);">$${(p.precio || 0).toLocaleString('es-CO')}</div>
                    ${variantsIndicator}
                </div>
                <div style="margin-top:12px; display:flex; justify-content:space-between; align-items:center;">
                    <span class="badge ${badgeClass}" style="font-size:9px;">${badgeText}</span>
                    <span class="mono" style="font-size:10px; color:var(--text-mut); font-weight:500;">${p.sku || ''}</span>
                </div>
            </div>
        `;
    }).join('');
}

function searchPOSProducts() {
    renderPOSProducts();
}

// Lógica del Carrito de Ventas
function addToCart(id, variantId = null) {
    const p = posProducts.find(x => x.id === id);
    if (!p) return;

    // Si tiene variantes y no se especificó una, seleccionar la primera por defecto para agilizar el flujo de venta rápida
    if (p.tiene_variantes && variantId === null) {
        if (p.variantes && p.variantes.length > 0) {
            variantId = p.variantes[0].id;
        }
    }

    const cartItem = cart.find(c => c.id === id && c.variantId === variantId);
    if (cartItem) {
        if (p.stock > 0 && cartItem.qty >= p.stock && !p.permite_stock_negativo) {
            Toast.warning('Límite de stock', `Stock máximo alcanzado para ${p.nombre}`);
            return;
        }
        cartItem.qty++;
    } else {
        let variantName = '';
        let sku = p.sku;
        let precio = p.precio;

        if (variantId !== null && p.variantes) {
            const v = p.variantes.find(x => x.id === variantId);
            if (v) {
                sku = v.sku || p.sku;
                precio = v.precio_override > 0 ? v.precio_override : p.precio;
                if (v.atributos) {
                    variantName = v.atributos.map(a => a.value).join(', ');
                }
            }
        }

        cart.push({
            id: p.id,
            nombre: p.nombre,
            precio: precio,
            sku: sku,
            stock: p.stock,
            controlaInventario: p.controla_inventario,
            permiteStockNegativo: p.permite_stock_negativo,
            variantId: variantId,
            variantInfo: variantName,
            qty: 1
        });
    }
    renderCart();
    playConfirmBeep();
}

function removeFromCart(id, variantId = null) {
    cart = cart.filter(c => !(c.id === id && c.variantId === variantId));
    renderCart();
}

function updateCartQty(id, variantId, delta) {
    const item = cart.find(c => c.id === id && c.variantId === variantId);
    if (!item) return;

    item.qty += delta;

    if (delta > 0 && item.controlaInventario && item.qty > item.stock && !item.permiteStockNegativo) {
        item.qty = item.stock;
        Toast.warning('Límite de stock', 'Stock máximo alcanzado');
        return;
    }

    if (item.qty <= 0) {
        removeFromCart(id, variantId);
    } else {
        renderCart();
    }
}

function renderCart() {
    const totalQty = cart.reduce((sum, item) => sum + item.qty, 0);
    document.getElementById('cart-count').textContent = `(${totalQty})`;

    const container = document.getElementById('cart-items');
    if (!cart.length) {
        container.innerHTML = `
            <div style="text-align: center; padding: 60px 20px; color: var(--text-mut);">
                <svg style="width: 48px; height: 48px; margin: 0 auto 12px auto; opacity: 0.3;" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M2.25 3h1.386c.51 0 .955.343 1.087.835l.383 1.437M7.5 14.25a3 3 0 00-3 3h15.75m-12.75-3h11.218c1.121-2.3 2.1-4.684 2.924-7.138a60.114 60.114 0 00-16.536-1.84M7.5 14.25L5.106 5.272M6 20.25a.75.75 0 11-1.5 0 .75.75 0 011.5 0zm12.75 0a.75.75 0 11-1.5 0 .75.75 0 011.5 0z"></path>
                </svg>
                Carrito vacío
            </div>
        `;
        document.getElementById('cart-subtotal').textContent = '$0';
        document.getElementById('cart-tax').textContent = '$0';
        document.getElementById('cart-total').textContent = '$0';
        return;
    }

    const total = cart.reduce((sum, item) => sum + (item.precio * item.qty), 0);
    const subtotal = total / 1.19; // IVA 19% colombiano
    const tax = total - subtotal;

    container.innerHTML = cart.map(item => {
        const varLabel = item.variantInfo ? `<div style="font-size:11px;color:var(--accent);margin-top:2px;">${item.variantInfo}</div>` : '';
        return `
            <div style="display:flex; justify-content:space-between; align-items:center; padding:10px 0; border-bottom:1px solid var(--border);">
                <div style="flex: 1; padding-right: 12px;">
                    <div style="font-weight:600; font-size:13px; color:var(--text);">${item.nombre}</div>
                    ${varLabel}
                    <div style="font-size:11px; color:var(--text-mut); margin-top:2px;">$${item.precio.toLocaleString('es-CO')} c/u</div>
                </div>
                <div style="display:flex; align-items:center; gap:8px;">
                    <div style="display:flex; align-items:center; border:1px solid var(--border); border-radius:6px; background:var(--bg-surf); overflow:hidden;">
                        <button class="action-btn" onclick="updateCartQty(${item.id}, ${item.variantId || 0}, -1)" style="width:24px; height:24px; padding:0; display:flex; align-items:center; justify-content:center; font-size:14px; border:none; background:transparent; cursor:pointer; color:var(--text-sec);">&minus;</button>
                        <span class="mono" style="min-width:24px; text-align:center; font-size:12px; font-weight:600;">${item.qty}</span>
                        <button class="action-btn" onclick="updateCartQty(${item.id}, ${item.variantId || 0}, 1)" style="width:24px; height:24px; padding:0; display:flex; align-items:center; justify-content:center; font-size:14px; border:none; background:transparent; cursor:pointer; color:var(--text-sec);">+</button>
                    </div>
                    <span class="mono" style="font-weight:600; font-size:13px; min-width:75px; text-align:right; color:var(--text);">$${(item.precio * item.qty).toLocaleString('es-CO')}</span>
                    <button class="action-btn" onclick="removeFromCart(${item.id}, ${item.variantId || 0})" style="color:var(--danger); margin-left:4px; display:flex; align-items:center; justify-content:center; width:24px; height:24px; border:none; background:transparent; cursor:pointer; font-size:16px;">&times;</button>
                </div>
            </div>
        `;
    }).join('');

    document.getElementById('cart-subtotal').textContent = `$${Math.round(subtotal).toLocaleString('es-CO')}`;
    document.getElementById('cart-tax').textContent = `$${Math.round(tax).toLocaleString('es-CO')}`;
    document.getElementById('cart-total').textContent = `$${Math.round(total).toLocaleString('es-CO')}`;
}

// Configuración de lector de códigos de barras (Escáner de teclado global)
function setupBarcodeScanner() {
    let barcodeBuffer = "";
    let lastKeyTime = Date.now();

    window.addEventListener("keydown", async (e) => {
        if (e.key === "F12") {
            e.preventDefault();
            const modal = document.getElementById('checkout-modal');
            if (modal && modal.style.display === 'flex') {
                submitPOSSale();
            } else {
                openCheckoutModal();
            }
            return;
        }

        const activeTag = document.activeElement.tagName;
        const isInput = activeTag === "INPUT" || activeTag === "TEXTAREA" || activeTag === "SELECT";

        const now = Date.now();
        const timeDiff = now - lastKeyTime;
        lastKeyTime = now;

        // Ignorar teclas de control de sistema y modificado
        if (e.key.length > 1 && e.key !== "Enter") {
            return;
        }

        // Si la diferencia entre teclas es menor a 40ms, se asume entrada por escáner de hardware
        if (timeDiff < 40) {
            if (e.key === "Enter") {
                if (barcodeBuffer.length > 2) {
                    e.preventDefault();
                    e.stopPropagation();
                    const code = barcodeBuffer;
                    barcodeBuffer = "";

                    // Limpiar el texto autocompletado en el input si estaba enfocado
                    if (isInput) {
                        const val = document.activeElement.value;
                        if (val.endsWith(code)) {
                            document.activeElement.value = val.slice(0, -code.length);
                        }
                        document.activeElement.blur();
                    }

                    await scanBarcodeByCode(code);
                }
            } else {
                barcodeBuffer += e.key;
            }
        } else {
            // Teclado manual lento, reiniciar buffer
            barcodeBuffer = e.key !== "Enter" ? e.key : "";
        }
    });
}

async function scanBarcodeByCode(code) {
    try {
        const r = await fetch(`${API_URL}/products/barcode/${code}`, {
            headers: { 'Authorization': 'Bearer ' + token }
        });
        if (r.ok) {
            const res = await r.json();
            const p = res.data;
            let variantId = null;

            // Identificar si corresponde a una variante específica
            if (p.variantes && p.variantes.length > 0) {
                const match = p.variantes.find(v => v.barcode === code || v.sku === code);
                if (match) variantId = match.id;
            }

            addToCart(p.id, variantId);
            Toast.success('Escaneado', `${p.nombre} agregado`);
        } else {
            Toast.error('Escáner', `Código "${code}" no encontrado`);
            playErrorBeep();
        }
    } catch (e) {
        console.error("Error en escaneo", e);
    }
}

// Efectos de Sonido
let audioCtx = null;
function playConfirmBeep() {
    try {
        if (!audioCtx) audioCtx = new (window.AudioContext || window.webkitAudioContext)();
        const osc = audioCtx.createOscillator();
        const gain = audioCtx.createGain();
        osc.type = "sine";
        osc.frequency.setValueAtTime(1000, audioCtx.currentTime);
        gain.gain.setValueAtTime(0.05, audioCtx.currentTime);
        osc.connect(gain);
        gain.connect(audioCtx.destination);
        osc.start();
        osc.stop(audioCtx.currentTime + 0.08);
    } catch (e) {}
}

function playErrorBeep() {
    try {
        if (!audioCtx) audioCtx = new (window.AudioContext || window.webkitAudioContext)();
        const osc = audioCtx.createOscillator();
        const gain = audioCtx.createGain();
        osc.type = "sawtooth";
        osc.frequency.setValueAtTime(250, audioCtx.currentTime);
        gain.gain.setValueAtTime(0.1, audioCtx.currentTime);
        osc.connect(gain);
        gain.connect(audioCtx.destination);
        osc.start();
        osc.stop(audioCtx.currentTime + 0.25);
    } catch (e) {}
}

function setupKeyboardShortcuts() {
    window.addEventListener("keydown", (e) => {
        if (e.key === "Escape") {
            closeCheckoutModal();
            closeNewCustomerModal();
        }

        // Shortcut modifiers for checkout modal
        const checkoutModal = document.getElementById('checkout-modal');
        if (checkoutModal && checkoutModal.style.display === 'flex') {
            if (e.altKey) {
                const key = e.key.toLowerCase();
                if (key === 'e') {
                    e.preventDefault();
                    selectPayMethod('efectivo');
                } else if (key === 't') {
                    e.preventDefault();
                    selectPayMethod('tarjeta');
                } else if (key === 'n') {
                    e.preventDefault();
                    selectPayMethod('transferencia');
                } else if (key === 'x') {
                    e.preventDefault();
                    setExactCash();
                }
            }
        }

        // Ctrl+Enter form submission
        if (e.ctrlKey && e.key === "Enter") {
            const custModal = document.getElementById('customer-modal');
            if (custModal && custModal.style.display === 'flex') {
                e.preventDefault();
                const form = document.getElementById('new-customer-form');
                if (typeof form.requestSubmit === 'function') {
                    form.requestSubmit();
                } else {
                    if (form.reportValidity()) {
                        submitNewCustomer(new Event('submit'));
                    }
                }
            } else if (checkoutModal && checkoutModal.style.display === 'flex') {
                e.preventDefault();
                submitPOSSale();
            }
        }
    });
}

// Búsqueda y Selección de Clientes
async function searchCustomers() {
    const q = document.getElementById('pos-customer-search').value.trim();
    const resultsDiv = document.getElementById('customer-search-results');

    if (q.length < 2) {
        resultsDiv.style.display = 'none';
        return;
    }

    try {
        const r = await fetch(`${API_URL}/customers?search=${encodeURIComponent(q)}&limit=5`, {
            headers: { 'Authorization': 'Bearer ' + token }
        });
        if (r.ok) {
            const res = await r.json();
            const list = res.data || [];

            if (list.length === 0) {
                resultsDiv.innerHTML = '<div style="padding:12px; font-size:12px; color:var(--text-mut); text-align:center;">Sin resultados</div>';
            } else {
                let html = `
                    <div style="display: grid; grid-template-columns: 100px 1fr 110px; padding: 8px 12px; font-weight: 600; font-size: 10px; color: var(--text-mut); border-bottom: 1px solid var(--border-strong); background: rgba(0,0,0,0.15); letter-spacing: 0.5px;">
                        <div>DOCUMENTO</div>
                        <div>NOMBRE</div>
                        <div style="text-align: right;">TELÉFONO</div>
                    </div>
                `;
                html += list.map(c => `
                    <div class="customer-result-row" onclick="selectCustomer(${c.id}, '${c.nombre.replace(/'/g, "\\'")}', '${c.cedula}')" style="display: grid; grid-template-columns: 100px 1fr 110px; padding: 10px 12px; border-bottom: 1px solid var(--border); cursor: pointer; font-size: 12px; transition: background 0.15s; align-items: center;" onmouseover="this.style.background='var(--bg-hover)';" onmouseout="this.style.background='transparent';">
                        <div class="mono" style="font-weight: 600; color: var(--accent);">${c.cedula}</div>
                        <div style="font-weight: 600; color: var(--text); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; padding-right: 8px;">${c.nombre}</div>
                        <div class="mono" style="text-align: right; color: var(--text-sec);">${c.telefono || '-'}</div>
                    </div>
                `).join('');
                resultsDiv.innerHTML = html;
            }
            resultsDiv.style.display = 'block';
        }
    } catch (e) {
        console.error("Error buscando clientes", e);
    }
}

function selectCustomer(id, nombre, cedula) {
    selectedCustomer = { id, nombre, cedula };
    document.getElementById('pos-customer-search').value = '';
    document.getElementById('customer-search-results').style.display = 'none';
    document.getElementById('customer-search-container').style.display = 'none';

    document.getElementById('sel-customer-name').textContent = nombre;
    document.getElementById('sel-customer-id').textContent = cedula;
    document.getElementById('selected-customer-box').style.display = 'flex';
}

function clearSelectedCustomer() {
    selectedCustomer = null;
    document.getElementById('selected-customer-box').style.display = 'none';
    document.getElementById('customer-search-container').style.display = 'block';
    document.getElementById('pos-customer-search').focus();
}

function openNewCustomerModal() {
    document.getElementById('customer-modal').style.display = 'flex';
    setTimeout(() => {
        const input = document.getElementById('cust-cedula');
        if (input) {
            input.focus();
            input.select();
        }
    }, 50);
}

function closeNewCustomerModal() {
    document.getElementById('customer-modal').style.display = 'none';
    document.getElementById('new-customer-form').reset();
}

async function submitNewCustomer(e) {
    e.preventDefault();

    const cedula = document.getElementById('cust-cedula').value.trim();
    const nombre = document.getElementById('cust-nombre').value.trim();
    const telefono = document.getElementById('cust-telefono').value.trim();
    const email = document.getElementById('cust-email').value.trim();

    try {
        const r = await fetch(`${API_URL}/customers`, {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + token,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ cedula, nombre, email, telefono })
        });

        if (r.ok) {
            const res = await r.json();
            const newCustomer = res.data;
            selectCustomer(newCustomer.id, newCustomer.nombre, newCustomer.cedula);
            closeNewCustomerModal();
            Toast.success('Cliente', 'Registrado y seleccionado');
        } else {
            const err = await r.json();
            Toast.error('Error', err.error || 'No se pudo registrar cliente');
        }
    } catch (e) {
        Toast.error('Error', 'Ocurrió un error al guardar');
    }
}

// Lógica de Cobro y Checkout
function openCheckoutModal() {
    if (cart.length === 0) {
        Toast.warning('Carrito vacío', 'Agrega productos antes de cobrar');
        return;
    }

    const total = cart.reduce((sum, item) => sum + (item.precio * item.qty), 0);
    document.getElementById('modal-total-pay').textContent = `$${total.toLocaleString('es-CO')}`;
    document.getElementById('checkout-modal').style.display = 'flex';

    selectPayMethod('efectivo');
    document.getElementById('pay-cash-received').value = '';
    document.getElementById('modal-change').textContent = '$0';
    document.getElementById('modal-change').style.color = 'var(--success)';
    
    setTimeout(() => {
        const input = document.getElementById('pay-cash-received');
        if (input) {
            input.focus();
            input.select();
        }
    }, 50);
}

function closeCheckoutModal() {
    document.getElementById('checkout-modal').style.display = 'none';
}

function selectPayMethod(method) {
    selectedPayMethod = method;

    document.querySelectorAll('.pay-method-btn').forEach(btn => {
        btn.classList.remove('active');
        btn.style.borderColor = 'var(--border)';
        btn.style.background = 'var(--bg-surf)';
        btn.style.color = 'var(--text-sec)';
    });

    const activeBtn = document.getElementById(`pay-method-${method}`);
    activeBtn.classList.add('active');
    activeBtn.style.borderColor = 'var(--accent)';
    activeBtn.style.background = 'var(--accent-dim)';
    activeBtn.style.color = 'var(--accent)';

    const calculatorFields = document.getElementById('cash-calculator-fields');

    if (method === 'efectivo') {
        calculatorFields.style.opacity = '1';
        calculatorFields.style.pointerEvents = 'auto';
        setTimeout(() => {
            const input = document.getElementById('pay-cash-received');
            if (input) {
                input.focus();
                input.select();
            }
        }, 50);
    } else {
        calculatorFields.style.opacity = '0.5';
        calculatorFields.style.pointerEvents = 'none';

        const total = cart.reduce((sum, item) => sum + (item.precio * item.qty), 0);
        document.getElementById('pay-cash-received').value = total.toLocaleString('es-CO');
        document.getElementById('modal-change').textContent = '$0';
        document.getElementById('modal-change').style.color = 'var(--text-mut)';
    }
}

function calculateChange() {
    const total = cart.reduce((sum, item) => sum + (item.precio * item.qty), 0);
    const inputVal = document.getElementById('pay-cash-received').value.replace(/[^0-9]/g, '');
    const received = parseFloat(inputVal) || 0;

    if (inputVal) {
        document.getElementById('pay-cash-received').value = received.toLocaleString('es-CO');
    }

    const change = received - total;
    const changeEl = document.getElementById('modal-change');

    if (change >= 0) {
        changeEl.textContent = `$${Math.round(change).toLocaleString('es-CO')}`;
        changeEl.style.color = 'var(--success)';
        document.getElementById('change-box').style.background = 'rgba(34, 197, 94, 0.08)';
        document.getElementById('change-box').style.borderColor = 'rgba(34, 197, 94, 0.15)';
    } else {
        changeEl.textContent = `Falta: $${Math.abs(Math.round(change)).toLocaleString('es-CO')}`;
        changeEl.style.color = 'var(--danger)';
        document.getElementById('change-box').style.background = 'rgba(239, 68, 68, 0.08)';
        document.getElementById('change-box').style.borderColor = 'rgba(239, 68, 68, 0.15)';
    }
}

function addQuickCash(amount) {
    const inputEl = document.getElementById('pay-cash-received');
    const current = parseFloat(inputEl.value.replace(/[^0-9]/g, '')) || 0;
    const newAmount = current + amount;

    inputEl.value = newAmount.toLocaleString('es-CO');
    calculateChange();
}

function setExactCash() {
    const total = cart.reduce((sum, item) => sum + (item.precio * item.qty), 0);
    document.getElementById('pay-cash-received').value = total.toLocaleString('es-CO');
    calculateChange();
}

async function submitPOSSale() {
    if (cart.length === 0) return;

    const total = cart.reduce((sum, item) => sum + (item.precio * item.qty), 0);
    const receivedVal = parseFloat(document.getElementById('pay-cash-received').value.replace(/[^0-9]/g, '')) || 0;

    if (selectedPayMethod === 'efectivo' && receivedVal < total) {
        Toast.warning('Pago incompleto', 'El monto recibido es menor al total a cobrar');
        return;
    }

    const items = cart.map(item => ({
        producto_id: item.id,
        variant_id: item.variantId,
        cantidad: item.qty,
        precio_unitario: item.precio
    }));

    const payload = {
        items: items,
        metodo_pago: selectedPayMethod,
        recibido: selectedPayMethod === 'efectivo' ? receivedVal : total,
        customer_id: selectedCustomer ? selectedCustomer.id : null
    };

    try {
        const r = await fetch(`${API_URL}/orders/pos`, {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + token,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(payload)
        });

        if (r.ok) {
            Toast.success('Venta registrada', 'Imprimiendo comprobante...');
            
            // Reiniciar estado de venta
            cart = [];
            selectedCustomer = null;
            clearSelectedCustomer();
            renderCart();
            closeCheckoutModal();

            // Recargar catálogo de productos para actualizar cantidades de stock en pantalla
            await loadPOSProducts();
        } else {
            const err = await r.json();
            Toast.error('Error', err.error || 'Ocurrió un error al registrar venta');
        }
    } catch (e) {
        Toast.error('Error', 'No hay conexión con el servidor');
    }
}


async function loadPOSSales() {
    const period = document.getElementById('sales-period')?.value || 'today';
    let fechaDesde = '';

    if (period === 'today') {
        const today = new Date();
        fechaDesde = today.toISOString().split('T')[0];
    } else if (period === 'week') {
        const date = new Date();
        date.setDate(date.getDate() - 7);
        fechaDesde = date.toISOString().split('T')[0];
    } else if (period === 'month') {
        const date = new Date();
        date.setMonth(date.getMonth() - 1);
        fechaDesde = date.toISOString().split('T')[0];
    }

    try {
        let url = `${API_URL}/orders?limit=30`;
        if (fechaDesde) {
            url += `&fecha_desde=${fechaDesde}`;
        }

        const r = await fetch(url, {
            headers: { 'Authorization': 'Bearer ' + token }
        });
        
        if (r.ok) {
            const res = await r.json();
            const list = res.data || [];
            renderPOSSales(list);
        } else {
            console.error("Error al cargar historial de ventas");
        }
    } catch (e) {
        console.error("Error de conexion al cargar ventas", e);
    }
}

function renderPOSSales(orders) {
    const tbody = document.getElementById('sales-tbody');
    if (!tbody) return;

    if (orders.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" style="text-align:center;padding:30px;color:var(--text-mut)">No hay ventas registradas</td></tr>';
        return;
    }

    tbody.innerHTML = orders.map(o => {
        let dateStr = '-';
        if (o.created_at) {
            const dateObj = new Date(o.created_at);
            if (!isNaN(dateObj.getTime())) {
                dateStr = dateObj.toLocaleTimeString('es-CO', { hour: '2-digit', minute: '2-digit' });
            }
        }
        
        const itemsQty = o.items ? o.items.reduce((sum, i) => sum + i.cantidad, 0) : 0;
        const totalFormatted = (o.total || 0).toLocaleString('es-CO');
        const payMethodStr = o.metodo_pago ? o.metodo_pago.charAt(0).toUpperCase() + o.metodo_pago.slice(1) : '-';

        return `
            <tr>
                <td style="padding:12px 16px; font-weight:600;" class="mono">#${o.id}</td>
                <td style="padding:12px 16px; color:var(--text-sec);">${dateStr}</td>
                <td style="padding:12px 16px; text-align:center;" class="mono">${itemsQty}</td>
                <td style="padding:12px 16px; font-weight:600;" class="mono">$${totalFormatted}</td>
                <td style="padding:12px 16px;"><span class="badge badge-green">${payMethodStr}</span></td>
            </tr>
        `;
    }).join('');
}

// Inicialización
function initPOS() {
    loadPOSProducts();
    loadPOSCategories();
    setupBarcodeScanner();
    setupKeyboardShortcuts();
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initPOS);
} else {
    initPOS();
}