/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

// KindredCard - Global JavaScript
(function() {
    'use strict';

    // Theme Management
    window.setTheme = function(theme) {
        // 1. Immediate UI Feedback
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem('theme', theme);

        // 3. Sync with Server
        fetch('/api/v1/user/preferences', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ 
                theme: theme
            })
        }).catch(err => console.error('Failed to save preferences:', err));

        if (window.showNotification) {
            showNotification(`Theme changed to ${theme}`, 'info');
        }
    };

    // Load saved theme immediately to prevent "flash"
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme) {
        document.documentElement.setAttribute('data-theme', savedTheme);
    }

    // Modal Utilities
    window.openModal = function(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.add('modal-open');
        }
    };

    window.closeModal = function(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.remove('modal-open');
        }
    };

    // Toast Notification System
    window.showNotification = function(message, type = 'info') {
        const container = document.getElementById('toastContainer');
        if (!container) return;

        const alertClasses = {
            'success': 'alert-success',
            'error': 'alert-error',
            'warning': 'alert-warning',
            'info': 'alert-info'
        };

        const icons = {
            'success': `<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`,
            'error': `<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`,
            'warning': `<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>`,
            'info': `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>`
        };

        const toast = document.createElement('div');
        toast.className = `alert ${alertClasses[type]} shadow-lg mb-2 animate-slide-in`;
        toast.innerHTML = `
            <div>
                ${icons[type]}
                <span>${message}</span>
            </div>
        `;

        container.appendChild(toast);

        // Auto-remove after 3 seconds
        setTimeout(() => {
            toast.style.opacity = '0';
            toast.style.transform = 'translateX(100%)';
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    };

    // Confirm Dialog
    window.confirmDialog = function(message, callback) {
        if (confirm(message)) {
            callback();
        }
    };

    // Fetch Helper with Error Handling
    window.apiRequest = async function(url, options = {}) {
        try {
            const response = await fetch(url, {
                ...options,
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                }
            });

            if (!response.ok) {
                const error = await response.json().catch(() => ({ error: 'Request failed' }));
                throw new Error(error.error || `HTTP ${response.status}`);
            }

            return await response.json();
        } catch (error) {
            console.error('API Request Error:', error);
            showNotification(error.message, 'error');
            throw error;
        }
    };

    // Delete Contact Helper
    window.deleteContact = async function(contactId) {
        try {
            await apiRequest(`/api/v1/contacts/${contactId}`, {
                method: 'DELETE'
            });
            showNotification('Contact deleted successfully', 'success');
            return true;
        } catch (error) {
            return false;
        }
    };

    // Format Date Helper
    window.formatDate = function(dateStr) {
        if (!dateStr) return '';
        const date = new Date(dateStr);
        return date.toLocaleDateString('en-US', { 
            year: 'numeric', 
            month: 'long', 
            day: 'numeric' 
        });
    };

    // Month Name Helper
    window.monthName = function(month) {
        const months = ['', 'January', 'February', 'March', 'April', 'May', 'June',
                       'July', 'August', 'September', 'October', 'November', 'December'];
        return months[month] || '';
    };

    // Deref Helper (for template compatibility)
    window.deref = function(value) {
        return value || 0;
    };

    // Keyboard Shortcuts
    document.addEventListener('keydown', function(e) {
        // Ctrl/Cmd + K: Focus search
        if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
            e.preventDefault();
            const searchInput = document.getElementById('searchInput');
            if (searchInput) {
                searchInput.focus();
            }
        }

        // Escape: Close modals
        if (e.key === 'Escape') {
            const openModals = document.querySelectorAll('.modal.modal-open');
            openModals.forEach(modal => {
                modal.classList.remove('modal-open');
            });
        }
    });

    // Auto-close alerts on click
    document.addEventListener('click', function(e) {
        if (e.target.closest('.alert')) {
            const alert = e.target.closest('.alert');
            alert.style.opacity = '0';
            alert.style.transform = 'translateX(100%)';
            setTimeout(() => alert.remove(), 300);
        }
    });

    // Initialize tooltips (if using)
    document.querySelectorAll('[data-tip]').forEach(element => {
        element.classList.add('tooltip');
    });

    // Loading State Helper
    window.setLoading = function(button, isLoading) {
        if (isLoading) {
            button.disabled = true;
            button.classList.add('loading');
        } else {
            button.disabled = false;
            button.classList.remove('loading');
        }
    };

    // Copy to Clipboard
    window.copyToClipboard = function(text, successMessage = 'Copied to clipboard') {
        navigator.clipboard.writeText(text).then(() => {
            showNotification(successMessage, 'success');
        }).catch(err => {
            console.error('Copy failed:', err);
            showNotification('Failed to copy', 'error');
        });
    };

    // Initialize page
    console.log('KindredCard initialized');
})();
