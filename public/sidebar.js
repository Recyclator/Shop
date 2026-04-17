/**
 * Sidebar Component - Componente reutilizable para todas las páginas
 * Uso: Incluir este script y llamar a Sidebar.render('page-name')
 */

const Sidebar = {
    // Configuración de navegación
    navItems: [
        {
            section: 'Principal',
            items: [
                { page: 'dashboard', label: 'Dashboard', icon: 'dashboard', url: '/dashboard.html' },
                { page: 'products', label: 'Productos', icon: 'products', url: '/products.html' },
                { page: 'categories', label: 'Categorías', icon: 'categories', url: '/categories.html' },
                { page: 'customers', label: 'Clientes', icon: 'customers', url: '/customers.html' },
                { page: 'layaways', label: 'Separados', icon: 'layaways', url: '/layaways.html' },
                { page: 'pos', label: 'Ventas', icon: 'pos', url: '/pos.html' }
            ]
        },
        {
            section: 'Sistema',
            items: [
                { page: 'users', label: 'Usuarios', icon: 'users', url: '/users.html' },
                { page: 'roles-permissions', label: 'Roles y Permisos', icon: 'roles', url: '/roles-permissions.html' },
                { page: 'profile', label: 'Perfil', icon: 'profile', url: '/profile.html' },
                { page: 'settings', label: 'Configuración', icon: 'settings', url: '/settings.html' }
            ]
        }
    ],

    // Iconos SVG
    icons: {
        dashboard: '<rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/>',
        products: '<path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>',
        categories: '<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>',
        customers: '<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>',
        layaways: '<rect x="1" y="4" width="22" height="16" rx="2" ry="2"/><line x1="1" y1="10" x2="23" y2="10"/>',
        pos: '<line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>',
        users: '<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>',
        roles: '<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>',
        profile: '<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>',
        settings: '<circle cx="12" cy="12" r="3"/><path d="M19.07 4.93a10 10 0 0 1 0 14.14M4.93 4.93a10 10 0 0 0 0 14.14"/>'
    },

    /**
     * Renderiza el sidebar completo
     * @param {string} currentPage - Nombre de la página activa (ej: 'dashboard', 'products')
     * @param {string} containerId - ID del contenedor donde se renderizará (default: 'sidebar-container')
     */
    render(currentPage, containerId = 'sidebar-container') {
        const container = document.getElementById(containerId);
        if (!container) {
            console.error('Sidebar container not found:', containerId);
            return;
        }

        let html = '<aside class="sidebar">';

        this.navItems.forEach(section => {
            html += `<div class="nav-label">${section.section}</div>`;
            
            section.items.forEach(item => {
                const isActive = item.page === currentPage ? ' active' : '';
                const icon = this.icons[item.icon] || this.icons.dashboard;
                
                html += `<div class="nav-item${isActive}" onclick="location.href='${item.url}'">`;
                html += `<svg fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">${icon}</svg>`;
                html += `${item.label}`;
                html += `</div>`;
            });
        });

        html += '</aside>';
        container.innerHTML = html;
    }
};

// Auto-detect current page and render sidebar when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    const sidebarContainer = document.getElementById('sidebar-container');
    if (sidebarContainer) {
        // Detect page name from URL or data attribute
        let currentPage = sidebarContainer.dataset.page;
        if (!currentPage) {
            // Fallback: extract from URL
            const path = window.location.pathname;
            const match = path.match(/\/([a-z-]+)\.html$/);
            currentPage = match ? match[1] : 'dashboard';
        }
        Sidebar.render(currentPage);
    }
});
