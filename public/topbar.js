/**
 * Topbar Component - Componente reutilizable para todas las páginas
 * Uso: Incluir este script y llamar a Topbar.render()
 */

const Topbar = {
    /**
     * Renderiza el topbar completo
     * @param {Object} options - Opciones de configuración
     * @param {string} options.searchPlaceholder - Placeholder del input de búsqueda
     * @param {string} options.logoLink - URL del logo (default: /dashboard.html)
     * @param {string} options.containerId - ID del contenedor (default: 'topbar-container')
     */
    render(options = {}) {
        const {
            searchPlaceholder = 'Buscar...',
            logoLink = '/dashboard.html',
            containerId = 'topbar-container'
        } = options;

        const container = document.getElementById(containerId);
        if (!container) {
            console.error('Topbar container not found:', containerId);
            return;
        }

        // Load user data
        const userStr = localStorage.getItem('nexora_user');
        let userName = 'Usuario';
        let userInitial = '?';
        
        if (userStr) {
            try {
                const user = JSON.parse(userStr);
                userName = user.nombre || user.email;
                userInitial = userName.charAt(0).toUpperCase();
            } catch (e) {}
        }

        const html = `
            <header class="topbar">
                <div class="logo" onclick="location.href='${logoLink}'">
                    <div class="logo-icon"><svg viewBox="0 0 16 16" fill="white"><path d="M3 3h4v4H3zM9 3h4v4H9zM3 9h4v4H3zM11 9l2 4-4-2 2-2z"/></svg></div>
                    <span>Nexora</span>
                </div>
                <div class="search">
                    <input type="text" placeholder="${searchPlaceholder}" id="search-input">
                </div>
                <div class="spacer"></div>
                <div class="user">
                    <div class="avatar" id="user-avatar" onclick="location.href='/profile.html'" style="cursor:pointer">${userInitial}</div>
                    <span id="user-name">${userName}</span>
                </div>
                <button class="logout-btn" onclick="Topbar.logout()">Salir</button>
            </header>
        `;

        container.innerHTML = html;
    },

    /**
     * Logout function
     */
    logout() {
        localStorage.removeItem('nexora_token');
        localStorage.removeItem('nexora_user');
        window.location.href = '/index.html';
    },

    /**
     * Update user info display
     */
    updateUserInfo() {
        const userStr = localStorage.getItem('nexora_user');
        if (userStr) {
            try {
                const user = JSON.parse(userStr);
                const nameEl = document.getElementById('user-name');
                const avatarEl = document.getElementById('user-avatar');
                
                if (nameEl) nameEl.textContent = user.nombre || user.email;
                if (avatarEl && user.nombre) {
                    avatarEl.textContent = user.nombre.charAt(0).toUpperCase();
                }
            } catch (e) {}
        }
    }
};

// Auto-render topbar when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    const topbarContainer = document.getElementById('topbar-container');
    if (topbarContainer) {
        const placeholder = topbarContainer.dataset.searchPlaceholder;
        Topbar.render({ searchPlaceholder: placeholder || 'Buscar...' });
    }
});
