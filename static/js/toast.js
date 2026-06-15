const Toast = {
    container: null,
    
    init() {
        if (this.container) return;
        
        const style = document.createElement('style');
        style.textContent = `
            .toast-container {
                position: fixed;
                top: 20px;
                right: 20px;
                z-index: 9999;
                display: flex;
                flex-direction: column;
                gap: 10px;
                max-width: 360px;
            }
            .toast {
                padding: 14px 18px;
                border-radius: 12px;
                background: var(--bg-elev, #1a1a1a);
                border: 1px solid var(--border, #333);
                color: var(--text, #fff);
                font-size: 14px;
                box-shadow: 0 8px 32px rgba(0,0,0,0.4);
                animation: toast-in 0.3s ease;
                display: flex;
                align-items: flex-start;
                gap: 12px;
            }
            .toast.toast-out {
                animation: toast-out 0.3s ease forwards;
            }
            .toast-icon {
                width: 20px;
                height: 20px;
                flex-shrink: 0;
                border-radius: 50%;
                display: flex;
                align-items: center;
                justify-content: center;
            }
            .toast-success .toast-icon { background: rgba(34,197,94,0.2); color: #22c55e; }
            .toast-error .toast-icon { background: rgba(239,68,68,0.2); color: #ef4444; }
            .toast-warning .toast-icon { background: rgba(251,191,36,0.2); color: #fbbf24; }
            .toast-info .toast-icon { background: rgba(59,130,246,0.2); color: #3b82f6; }
            .toast-content { flex: 1; }
            .toast-title { font-weight: 600; margin-bottom: 2px; }
            .toast-message { opacity: 0.85; font-size: 13px; }
            .toast-close {
                background: none;
                border: none;
                color: var(--text-mut, #666);
                cursor: pointer;
                padding: 0;
                font-size: 18px;
                line-height: 1;
            }
            .toast-close:hover { color: var(--text, #fff); }
            @keyframes toast-in {
                from { transform: translateX(100%); opacity: 0; }
                to { transform: translateX(0); opacity: 1; }
            }
            @keyframes toast-out {
                from { transform: translateX(0); opacity: 1; }
                to { transform: translateX(100%); opacity: 0; }
            }
        `;
        document.head.appendChild(style);
        
        this.container = document.createElement('div');
        this.container.className = 'toast-container';
        document.body.appendChild(this.container);
    },
    
    show(type, title, message, duration = 4000) {
        this.init();
        
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        
        const icons = {
            success: '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"></polyline></svg>',
            error: '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>',
            warning: '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><path d="M12 9v4M12 17h.01"></path><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path></svg>',
            info: '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>'
        };
        
        toast.innerHTML = `
            <div class="toast-icon">${icons[type]}</div>
            <div class="toast-content">
                <div class="toast-title">${title}</div>
                <div class="toast-message">${message}</div>
            </div>
            <button class="toast-close" onclick="this.parentElement.remove()">&times;</button>
        `;
        
        this.container.appendChild(toast);
        
        if (duration > 0) {
            setTimeout(() => {
                toast.classList.add('toast-out');
                setTimeout(() => toast.remove(), 300);
            }, duration);
        }
        
        return toast;
    },
    
    success(title, message) { return this.show('success', title, message); },
    error(title, message) { return this.show('error', title, message); },
    warning(title, message) { return this.show('warning', title, message); },
    info(title, message) { return this.show('info', title, message); }
};

// Backwards compatibility
function showAlert(type, title, message) {
    Toast.show(type, title, message);
}

function hideAlert() {
    // No longer needed with toast
}
