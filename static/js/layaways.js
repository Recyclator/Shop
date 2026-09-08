function getAuthToken() { return window.token || localStorage.getItem('nexora_token') || ''; }
function getBaseApiUrl() { return window.API_URL || '/api'; }

function esc(str) {
    if (str === null || str === undefined) return '';
    return String(str)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}

let layCurrentPage = 1, layCurrentSearch = '', layCurrentStatus = '', layCurrentId = null;
function layFormat(n) { return (n||0).toLocaleString('es-CO'); }

async function loadLayaways(page=1) {
    layCurrentPage = page;
    let url = `${getBaseApiUrl()}/layaways?page=${page}&limit=20`;
    if (layCurrentSearch) url += '&search=' + encodeURIComponent(layCurrentSearch);
    if (layCurrentStatus) url += '&estado=' + layCurrentStatus;
    try {
        const res = await fetch(url, { headers: { 'Authorization': 'Bearer ' + getAuthToken() } });
        if (res.ok) {
            const d = await res.json();
            renderLayaways(d.data||[]);
            document.getElementById('total-layaways').textContent = `${d.meta?.total||0} separados`;
        } else {
            const tb = document.getElementById('layaways-table');
            if (tb) {
                tb.innerHTML = '<tr><td colspan="8" style="text-align:center;padding:40px;color:var(--text-mut)">Error al cargar separados o sin permisos</td></tr>';
            }
        }
    } catch(e) {
        Toast.error('Error','Error de conexión');
        const tb = document.getElementById('layaways-table');
        if (tb) {
            tb.innerHTML = '<tr><td colspan="8" style="text-align:center;padding:40px;color:var(--text-mut)">Error de conexión al cargar separados</td></tr>';
        }
    }
}

function renderLayaways(list) {
    const tb = document.getElementById('layaways-table');
    if (!tb) return;
    if (!list || !list.length) { tb.innerHTML = '<tr><td colspan="8" style="text-align:center;padding:40px;color:var(--text-mut)">Sin separados</td></tr>'; return; }
    tb.innerHTML = list.map(l => {
        const ab = (l.precio_total || 0) - (l.saldo_pendiente || 0);
        const bc = l.estado==='activo'?'badge-green':l.estado==='pagado'?'badge-blue':l.estado==='vencido'?'badge-red':'badge-gray';
        
        let dateStr = '-';
        if (l.fecha_vencimiento) {
            const dObj = new Date(l.fecha_vencimiento);
            if (!isNaN(dObj.getTime()) && dObj.getFullYear() > 1970) {
                dateStr = dObj.toLocaleDateString('es-CO');
            }
        }

        const layId = Number(l.id);
        const cancelBtn = (l.estado === 'activo' || l.estado === 'active') ? `
            <button class="action-btn" onclick="cancelLayaway(${layId})" title="Cancelar"><svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>
        ` : '';

        return `<tr>
            <td style="font-weight:500;color:var(--text)">${esc(l.customer?.nombre||'-')}</td>
            <td>${esc(l.product?.nombre||'-')}</td>
            <td class="mono">$${layFormat(l.precio_total)}</td>
            <td class="mono" style="color:var(--success)">$${layFormat(ab)}</td>
            <td class="mono" style="color:${l.saldo_pendiente>0?'var(--warning)':'var(--success)'}">$${layFormat(l.saldo_pendiente)}</td>
            <td>${esc(dateStr)}</td>
            <td><span class="badge ${bc}">${esc(l.estado)}</span></td>
            <td><div class="actions">
                <button class="action-btn" onclick="openPaymentsModal(${layId})" title="Abonos"><svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><rect x="1" y="4" width="22" height="16" rx="2" ry="2"/><line x1="1" y1="10" x2="23" y2="10"/></svg></button>
                ${cancelBtn}
            </div></td></tr>`;
    }).join('');
}

async function loadLayCustomers() { try { const res = await fetch(`${getBaseApiUrl()}/customers?limit=1000`,{headers:{'Authorization':'Bearer '+getAuthToken()}}); if(res.ok){const c=(await res.json()).data||[];document.getElementById('lay-customer').innerHTML='<option value="">Seleccionar...</option>'+c.map(x=>`<option value="${Number(x.id)}">${esc(x.nombre)} (${esc(x.cedula)})</option>`).join('');}} catch(e){} }
async function loadLayProducts() { try { const res = await fetch(`${getBaseApiUrl()}/products?limit=1000`,{headers:{'Authorization':'Bearer '+getAuthToken()}}); if(res.ok){const p=(await res.json()).data||[];document.getElementById('lay-product').innerHTML='<option value="">Seleccionar...</option>'+p.map(x=>`<option value="${Number(x.id)}">${esc(x.nombre)} - $${layFormat(x.precio)}</option>`).join('');}} catch(e){} }
function openCreateModal() { loadLayCustomers(); loadLayProducts(); document.getElementById('layaway-form').reset(); document.getElementById('layaway-modal').classList.add('active'); }
function closeModal() { document.getElementById('layaway-modal').classList.remove('active'); }
async function saveLayaway() {
    if (!FormValidator.validateForm('layaway-form')) return;
    const b = { customer_id:parseInt(document.getElementById('lay-customer').value), product_id:parseInt(document.getElementById('lay-product').value), cantidad:parseInt(document.getElementById('lay-cantidad').value), abono_inicial:parseFloat(document.getElementById('lay-abono').value) };
    try{const res=await fetch(`${getBaseApiUrl()}/layaways`,{method:'POST',headers:{'Authorization':'Bearer '+getAuthToken(),'Content-Type':'application/json'},body:JSON.stringify(b)});const d=await res.json();if(res.ok){Toast.success('Éxito','Separado creado');closeModal();loadLayaways(layCurrentPage);}else{Toast.error('Error',d.error?.message||'Error');}}catch(e){Toast.error('Error','Error de conexión');}
}
async function openPaymentsModal(id) {
    layCurrentId = id;
    try {
        const res = await fetch(`${getBaseApiUrl()}/layaways/${id}/payments`,{headers:{'Authorization':'Bearer '+getAuthToken()}});
        if(res.ok){
            const d = await res.json(), lay=d.separado, ab=d.abonos||[];
            document.getElementById('balance-amount').textContent = '$'+layFormat(lay.saldo_pendiente);
            document.getElementById('payment-details').innerHTML = `<div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;margin-bottom:16px;"><div><div style="font-size:11px;color:var(--text-mut)">Cliente</div><div style="font-weight:500;color:var(--text)">${esc(lay.customer?.nombre||'-')}</div></div><div><div style="font-size:11px;color:var(--text-mut)">Producto</div><div style="font-weight:500;color:var(--text)">${esc(lay.product?.nombre||'-')}</div></div><div><div style="font-size:11px;color:var(--text-mut)">Total</div><div class="mono">$${layFormat(lay.precio_total)}</div></div><div><div style="font-size:11px;color:var(--text-mut)">Abono Inicial</div><div class="mono" style="color:var(--success)">$${layFormat(lay.abono_inicial)}</div></div></div>`;
            document.getElementById('payments-list').innerHTML = ab.length===0?'<div style="text-align:center;padding:20px;color:var(--text-mut)">Sin abonos</div>':ab.map(a=>{
                let pDate = '-';
                if (a.fecha_pago) {
                    const dObj = new Date(a.fecha_pago);
                    if (!isNaN(dObj.getTime())) pDate = dObj.toLocaleString('es-CO');
                }
                return `<div style="padding:12px;border-bottom:1px solid var(--border);display:flex;justify-content:space-between;"><div><div class="mono" style="color:var(--success)">+$${layFormat(a.monto)}</div><div style="font-size:12px;color:var(--text-mut)">${esc(pDate)}</div></div></div>`;
            }).join('');
            document.getElementById('payment-amount').value = '';
            document.getElementById('payments-modal').classList.add('active');
        }
    } catch(e) { Toast.error('Error','No se pudo cargar'); }
}
function closePaymentsModal() { document.getElementById('payments-modal').classList.remove('active'); }
async function addPayment() {
    if (!FormValidator.validateForm('payment-form')) return;
    const m = parseFloat(document.getElementById('payment-amount').value);
    try{const res=await fetch(`${getBaseApiUrl()}/layaways/${layCurrentId}/payments`,{method:'POST',headers:{'Authorization':'Bearer '+getAuthToken(),'Content-Type':'application/json'},body:JSON.stringify({monto:m})});const d=await res.json();if(res.ok){Toast.success('Éxito',d.message||'Abono registrado');openPaymentsModal(layCurrentId);loadLayaways(layCurrentPage);}else{Toast.error('Error',d.error?.message||'Error');}}catch(e){Toast.error('Error','Error de conexión');}
}
async function cancelLayaway(id) { if(!confirm('¿Cancelar este separado?'))return; try{const res=await fetch(`${getBaseApiUrl()}/layaways/${id}/cancel`,{method:'POST',headers:{'Authorization':'Bearer '+getAuthToken()}});if(res.ok){Toast.success('Éxito','Separado cancelado');loadLayaways(layCurrentPage);}else{Toast.error('Error',(await res.json()).error?.message||'Error');}}catch(e){} }
async function checkExpired() { try{const res=await fetch(`${getBaseApiUrl()}/layaways/check-expired`,{method:'POST',headers:{'Authorization':'Bearer '+getAuthToken()}});if(res.ok){const d=await res.json();Toast.info('Verificación',`${d.data?.separados_vencidos||0} vencidos`);loadLayaways(layCurrentPage);}}catch(e){} }

function initLayaways() {
    const searchFilter = document.getElementById('search-filter');
    const statusFilter = document.getElementById('status-filter');
    if (searchFilter) {
        searchFilter.addEventListener('keypress', e => { if (e.key === 'Enter') { layCurrentSearch = e.target.value; loadLayaways(1); } });
    }
    if (statusFilter) {
        statusFilter.addEventListener('change', e => { layCurrentStatus = e.target.value; loadLayaways(1); });
    }
    loadLayaways(1);
}

window.loadLayaways = loadLayaways;
window.saveLayaway = saveLayaway;
window.openPaymentsModal = openPaymentsModal;
window.addPayment = addPayment;
window.cancelLayaway = cancelLayaway;
window.checkExpired = checkExpired;
window.openCreateModal = openCreateModal;
window.closeModal = closeModal;
window.closePaymentsModal = closePaymentsModal;

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initLayaways);
} else {
    initLayaways();
}
