let layCurrentPage = 1, layCurrentSearch = '', layCurrentStatus = '', layCurrentId = null;
function layFormat(n) { return (n||0).toLocaleString('es-CO'); }
async function loadLayaways(page=1) {
    layCurrentPage = page;
    let url = `${API_URL}/layaways?page=${page}&limit=20`;
    if (layCurrentSearch) url += '&search=' + encodeURIComponent(layCurrentSearch);
    if (layCurrentStatus) url += '&estado=' + layCurrentStatus;
    try {
        const res = await fetch(url, { headers: { 'Authorization': 'Bearer ' + token } });
        if (res.ok) { const d = await res.json(); renderLayaways(d.data||[]); document.getElementById('total-layaways').textContent = `${d.meta?.total||0} separados`; }
    } catch(e) { Toast.error('Error','Error de conexión'); }
}
function renderLayaways(list) {
    const tb = document.getElementById('layaways-table');
    if (!list.length) { tb.innerHTML = '<tr><td colspan="8" style="text-align:center;padding:40px;color:var(--text-mut)">Sin separados</td></tr>'; return; }
    tb.innerHTML = list.map(l => {
        const ab = l.precio_total - l.saldo_pendiente;
        const bc = l.estado==='activo'?'badge-green':l.estado==='pagado'?'badge-blue':l.estado==='vencido'?'badge-red':'badge-gray';
        return `<tr>
            <td style="font-weight:500;color:var(--text)">${l.customer?.nombre||'-'}</td>
            <td>${l.product?.nombre||'-'}</td>
            <td class="mono">$${layFormat(l.precio_total)}</td>
            <td class="mono" style="color:var(--success)">$${layFormat(ab)}</td>
            <td class="mono" style="color:${l.saldo_pendiente>0?'var(--warning)':'var(--success)'}">$${layFormat(l.saldo_pendiente)}</td>
            <td>${new Date(l.fecha_vencimiento).toLocaleDateString('es-CO')}</td>
            <td><span class="badge ${bc}">${l.estado}</span></td>
            <td><div class="actions">
                <button class="action-btn" onclick="openPaymentsModal(${l.id})" title="Abonos"><svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><rect x="1" y="4" width="22" height="16" rx="2" ry="2"/><line x1="1" y1="10" x2="23" y2="10"/></svg></button>
                ${l.estado==='active'?`<button class="action-btn" onclick="cancelLayaway(${l.id})" title="Cancelar"><svg width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>`:''}
            </div></td></tr>`;
    }).join('');
}
async function loadLayCustomers() { try { const res = await fetch(`${API_URL}/customers?limit=1000`,{headers:{'Authorization':'Bearer '+token}}); if(res.ok){const c=(await res.json()).data||[];document.getElementById('lay-customer').innerHTML='<option value="">Seleccionar...</option>'+c.map(x=>`<option value="${x.id}">${x.nombre} (${x.cedula})</option>`).join('');}} catch(e){} }
async function loadLayProducts() { try { const res = await fetch(`${API_URL}/products?limit=1000`,{headers:{'Authorization':'Bearer '+token}}); if(res.ok){const p=(await res.json()).data||[];document.getElementById('lay-product').innerHTML='<option value="">Seleccionar...</option>'+p.map(x=>`<option value="${x.id}">${x.nombre} - $${layFormat(x.precio)}</option>`).join('');}} catch(e){} }
function openCreateModal() { loadLayCustomers(); loadLayProducts(); document.getElementById('layaway-form').reset(); document.getElementById('layaway-modal').classList.add('active'); }
function closeModal() { document.getElementById('layaway-modal').classList.remove('active'); }
async function saveLayaway() {
    if (!FormValidator.validateForm('layaway-form')) return;
    const b = { customer_id:parseInt(document.getElementById('lay-customer').value), product_id:parseInt(document.getElementById('lay-product').value), cantidad:parseInt(document.getElementById('lay-cantidad').value), abono_inicial:parseFloat(document.getElementById('lay-abono').value) };
    try{const res=await fetch(`${API_URL}/layaways`,{method:'POST',headers:{'Authorization':'Bearer '+token,'Content-Type':'application/json'},body:JSON.stringify(b)});const d=await res.json();if(res.ok){Toast.success('Éxito','Separado creado');closeModal();loadLayaways(layCurrentPage);}else{Toast.error('Error',d.error?.message||'Error');}}catch(e){Toast.error('Error','Error de conexión');}
}
async function openPaymentsModal(id) {
    layCurrentId = id;
    try {
        const res = await fetch(`${API_URL}/layaways/${id}/payments`,{headers:{'Authorization':'Bearer '+token}});
        if(res.ok){
            const d = await res.json(), lay=d.separado, ab=d.abonos||[];
            document.getElementById('balance-amount').textContent = '$'+layFormat(lay.saldo_pendiente);
            document.getElementById('payment-details').innerHTML = `<div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;margin-bottom:16px;"><div><div style="font-size:11px;color:var(--text-mut)">Cliente</div><div style="font-weight:500;color:var(--text)">${lay.customer?.nombre||'-'}</div></div><div><div style="font-size:11px;color:var(--text-mut)">Producto</div><div style="font-weight:500;color:var(--text)">${lay.product?.nombre||'-'}</div></div><div><div style="font-size:11px;color:var(--text-mut)">Total</div><div class="mono">$${layFormat(lay.precio_total)}</div></div><div><div style="font-size:11px;color:var(--text-mut)">Abono Inicial</div><div class="mono" style="color:var(--success)">$${layFormat(lay.abono_inicial)}</div></div></div>`;
            document.getElementById('payments-list').innerHTML = ab.length===0?'<div style="text-align:center;padding:20px;color:var(--text-mut)">Sin abonos</div>':ab.map(a=>`<div style="padding:12px;border-bottom:1px solid var(--border);display:flex;justify-content:space-between;"><div><div class="mono" style="color:var(--success)">+$${layFormat(a.monto)}</div><div style="font-size:12px;color:var(--text-mut)">${new Date(a.fecha_pago).toLocaleString('es-CO')}</div></div></div>`).join('');
            document.getElementById('payment-amount').value = '';
            document.getElementById('payments-modal').classList.add('active');
        }
    } catch(e) { Toast.error('Error','No se pudo cargar'); }
}
function closePaymentsModal() { document.getElementById('payments-modal').classList.remove('active'); }
async function addPayment() {
    if (!FormValidator.validateForm('payment-form')) return;
    const m = parseFloat(document.getElementById('payment-amount').value);
    try{const res=await fetch(`${API_URL}/layaways/${layCurrentId}/payments`,{method:'POST',headers:{'Authorization':'Bearer '+token,'Content-Type':'application/json'},body:JSON.stringify({monto:m})});const d=await res.json();if(res.ok){Toast.success('Éxito',d.message||'Abono registrado');openPaymentsModal(layCurrentId);loadLayaways(layCurrentPage);}else{Toast.error('Error',d.error?.message||'Error');}}catch(e){Toast.error('Error','Error de conexión');}
}
async function cancelLayaway(id) { if(!confirm('¿Cancelar este separado?'))return; try{const res=await fetch(`${API_URL}/layaways/${id}/cancel`,{method:'POST',headers:{'Authorization':'Bearer '+token}});if(res.ok){Toast.success('Éxito','Separado cancelado');loadLayaways(layCurrentPage);}else{Toast.error('Error',(await res.json()).error?.message||'Error');}}catch(e){} }
async function checkExpired() { try{const res=await fetch(`${API_URL}/layaways/check-expired`,{method:'POST',headers:{'Authorization':'Bearer '+token}});if(res.ok){const d=await res.json();Toast.info('Verificación',`${d.data?.separados_vencidos||0} vencidos`);loadLayaways(layCurrentPage);}}catch(e){} }
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

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initLayaways);
} else {
    initLayaways();
}
