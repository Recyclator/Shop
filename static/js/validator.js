/**
 * FormValidator - A premium real-time form validation utility for Nexora
 */
const FormValidator = {
    // Default messages in Spanish
    messages: {
        required: "Este campo es obligatorio.",
        email: "Ingresa un correo electrónico válido.",
        minlength: "Debe tener al menos {minlength} caracteres.",
        maxlength: "No puede tener más de {maxlength} caracteres.",
        min: "El valor mínimo es {min}.",
        max: "El valor máximo es {max}.",
        match: "El campo no coincide.",
        pattern: "El formato no es válido."
    },

    init() {
        // Find all forms in the document
        document.querySelectorAll('form').forEach(form => {
            this.bindEvents(form);
        });

        // Use MutationObserver to bind events to dynamic forms (e.g. modals opening/rendered dynamically)
        const observer = new MutationObserver((mutations) => {
            mutations.forEach((mutation) => {
                mutation.addedNodes.forEach((node) => {
                    if (node.nodeType === Node.ELEMENT_NODE) {
                        if (node.tagName === 'FORM') {
                            this.bindEvents(node);
                        } else {
                            node.querySelectorAll('form').forEach(form => this.bindEvents(form));
                        }
                    }
                });
            });
        });
        observer.observe(document.body, { childList: true, subtree: true });
    },

    bindEvents(form) {
        if (form.dataset.validatorBound) return;
        form.dataset.validatorBound = "true";

        // Intercept standard submit events
        form.addEventListener('submit', (e) => {
            const isValid = this.validateForm(form);
            if (!isValid) {
                e.preventDefault();
                e.stopPropagation();
            }
        });

        // Clear error states on form reset
        form.addEventListener('reset', () => {
            // Delay slightly to allow values to clear first
            setTimeout(() => {
                const inputs = form.querySelectorAll('input, select, textarea');
                inputs.forEach(input => {
                    input.classList.remove('is-touched', 'is-invalid', 'is-valid');
                    const errorEl = input.parentNode.querySelector('.validation-error');
                    if (errorEl) {
                        errorEl.remove();
                    }
                });
            }, 0);
        });

        // Listen for blur and input/change events on fields
        const inputs = form.querySelectorAll('input, select, textarea');
        inputs.forEach(input => {
            if (input.type === 'hidden' || input.type === 'submit' || input.type === 'button') return;

            input.addEventListener('blur', () => {
                input.classList.add('is-touched');
                this.validateField(input);
            });

            input.addEventListener('input', () => {
                if (input.classList.contains('is-touched')) {
                    this.validateField(input);
                }
            });

            input.addEventListener('change', () => {
                if (input.classList.contains('is-touched')) {
                    this.validateField(input);
                }
            });
        });
    },

    validateField(input) {
        // Ignore hidden fields or fields inside hidden wrappers
        if (input.type === 'hidden' || input.disabled || input.closest('.hidden')) {
            this.clearError(input);
            return true;
        }

        const value = input.value.trim();
        let isValid = true;
        let errorCode = null;
        let errorParams = {};

        // 1. Required check
        if (input.hasAttribute('required') && value === '') {
            isValid = false;
            errorCode = 'required';
        }

        // 2. Email check
        if (isValid && input.type === 'email' && value !== '') {
            const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
            if (!emailRegex.test(value)) {
                isValid = false;
                errorCode = 'email';
            }
        }

        // 3. Minlength check
        if (isValid && input.hasAttribute('minlength') && value !== '') {
            const minLength = parseInt(input.getAttribute('minlength'));
            if (value.length < minLength) {
                isValid = false;
                errorCode = 'minlength';
                errorParams = { minlength: minLength };
            }
        }

        // 4. Maxlength check
        if (isValid && input.hasAttribute('maxlength') && value !== '') {
            const maxLength = parseInt(input.getAttribute('maxlength'));
            if (value.length > maxLength) {
                isValid = false;
                errorCode = 'maxlength';
                errorParams = { maxlength: maxLength };
            }
        }

        // 5. Min check
        if (isValid && input.hasAttribute('min') && value !== '') {
            const min = parseFloat(input.getAttribute('min'));
            const numVal = parseFloat(value);
            if (!isNaN(numVal) && numVal < min) {
                isValid = false;
                errorCode = 'min';
                errorParams = { min: min };
            }
        }

        // 6. Max check
        if (isValid && input.hasAttribute('max') && value !== '') {
            const max = parseFloat(input.getAttribute('max'));
            const numVal = parseFloat(value);
            if (!isNaN(numVal) && numVal > max) {
                isValid = false;
                errorCode = 'max';
                errorParams = { max: max };
            }
        }

        // 7. Match check (compara con otro campo por ID, ej. data-match="pwd-new")
        if (isValid && input.dataset.match && value !== '') {
            const targetId = input.dataset.match;
            const targetInput = document.getElementById(targetId);
            if (targetInput && targetInput.value.trim() !== value) {
                isValid = false;
                errorCode = 'match';
                if (targetId.includes('pass') || targetId.includes('pwd') || input.type === 'password') {
                    errorCode = 'passwordMatch';
                }
            }
        }

        // 8. Pattern check
        if (isValid && input.hasAttribute('pattern') && value !== '') {
            const pattern = new RegExp(input.getAttribute('pattern'));
            if (!pattern.test(value)) {
                isValid = false;
                errorCode = 'pattern';
            }
        }

        if (isValid) {
            this.clearError(input);
            return true;
        } else {
            let message = this.getErrorMessage(input, errorCode, errorParams);
            this.showError(input, message);
            return false;
        }
    },

    getErrorMessage(input, errorCode, params) {
        // Custom message directly defined on target
        // e.g. data-error-required="Nombre es obligatorio"
        const pascalErrorCode = errorCode.charAt(0).toUpperCase() + errorCode.slice(1);
        const customMessage = input.dataset[`error${pascalErrorCode}`] || input.dataset.errorMessage;
        if (customMessage) return customMessage;

        // Custom match messages
        if (errorCode === 'passwordMatch') {
            return "Las contraseñas no coinciden.";
        }

        let message = this.messages[errorCode] || "Campo no válido.";
        
        // Interpolate parameters
        for (const [key, val] of Object.entries(params)) {
            message = message.replace(`{${key}}`, val);
        }
        return message;
    },

    showError(input, message) {
        input.classList.remove('is-valid');
        input.classList.add('is-invalid');

        let errorEl = input.parentNode.querySelector('.validation-error');
        if (!errorEl) {
            errorEl = document.createElement('span');
            errorEl.className = 'validation-error';
            input.parentNode.appendChild(errorEl);
        }
        errorEl.textContent = message;
    },

    clearError(input) {
        input.classList.remove('is-invalid');
        if (input.value.trim() !== '' && (input.hasAttribute('required') || input.type === 'email')) {
            input.classList.add('is-valid');
        } else {
            input.classList.remove('is-valid');
        }

        const errorEl = input.parentNode.querySelector('.validation-error');
        if (errorEl) {
            errorEl.remove();
        }
    },

    validateForm(formElementOrId) {
        const form = typeof formElementOrId === 'string' ? document.getElementById(formElementOrId) : formElementOrId;
        if (!form) return true;

        this.bindEvents(form);

        const inputs = form.querySelectorAll('input, select, textarea');
        let isFormValid = true;
        let firstInvalidField = null;

        inputs.forEach(input => {
            if (input.type === 'hidden' || input.type === 'submit' || input.type === 'button') return;
            
            input.classList.add('is-touched');
            const isFieldValid = this.validateField(input);
            if (!isFieldValid) {
                isFormValid = false;
                if (!firstInvalidField) {
                    firstInvalidField = input;
                }
            }
        });

        if (firstInvalidField) {
            firstInvalidField.focus();
        }

        return isFormValid;
    }
};

window.FormValidator = FormValidator;

// Initialize on DOM load
document.addEventListener('DOMContentLoaded', () => {
    FormValidator.init();
});
