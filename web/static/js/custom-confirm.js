/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

// KindredCard - Replaces the native browser confirm() with a DaisyUI modal
(function() {
    'use strict';
    
    // Store the original confirm function
    const originalConfirm = window.confirm;
    
    /**
     * Custom confirm function that shows a modal instead of native dialog
     * @param {string} message - The message to display
     * @param {string} title - Optional title (defaults to "Confirm Action")
     * @returns {Promise<boolean>} - Resolves to true if confirmed, false if cancelled
     */
    window.customConfirm = function(message, title = 'Confirm Action') {
        return new Promise((resolve) => {
            const modal = document.getElementById('confirmModal');
            const titleEl = document.getElementById('confirmTitle');
            const messageEl = document.getElementById('confirmMessage');
            const okBtn = document.getElementById('confirmOk');
            const cancelBtn = document.getElementById('confirmCancel');
            
            if (!modal) {
                console.warn('Confirm modal not found, falling back to native confirm');
                resolve(originalConfirm(message));
                return;
            }
            
            // Set content
            titleEl.textContent = title;
            messageEl.textContent = message;
            
            // Remove old listeners
            const newOkBtn = okBtn.cloneNode(true);
            const newCancelBtn = cancelBtn.cloneNode(true);
            okBtn.parentNode.replaceChild(newOkBtn, okBtn);
            cancelBtn.parentNode.replaceChild(newCancelBtn, cancelBtn);
            
            // Add new listeners
            newOkBtn.addEventListener('click', () => {
                modal.close();
                resolve(true);
            });
            
            newCancelBtn.addEventListener('click', () => {
                modal.close();
                resolve(false);
            });
            
            // Handle backdrop click
            modal.addEventListener('close', () => {
                resolve(false);
            }, { once: true });
            
            // Show modal
            modal.showModal();
        });
    };
    
    /**
     * Override window.confirm to use custom modal
     * Note: This only works with async/await or promises
     */
    window.confirm = function(message) {
        console.warn('Synchronous confirm() called. Use customConfirm() with async/await for better UX');
        return originalConfirm(message);
    };
    
    /**
     * Attach to form onsubmit handlers that use confirm()
     * Example: <form onsubmit="return confirm('Are you sure?')">
     * Converts to: <form data-confirm="Are you sure?">
     */
    document.addEventListener('DOMContentLoaded', () => {
        // Handle forms with data-confirm attribute
        document.addEventListener('submit', async (e) => {
            const form = e.target;
            const confirmMessage = form.dataset.confirm;
            const confirmTitle = form.dataset.confirmTitle;
            
            if (confirmMessage) {
                e.preventDefault();
                
                const confirmed = await window.customConfirm(confirmMessage, confirmTitle);
                
                if (confirmed) {
                    // Remove the data-confirm attribute to prevent infinite loop
                    form.removeAttribute('data-confirm');
                    form.submit();
                }
            }
        });
        
        // Handle links with data-confirm attribute
        document.addEventListener('click', async (e) => {
            const link = e.target.closest('a[data-confirm]');
            
            if (link) {
                e.preventDefault();
                
                const confirmMessage = link.dataset.confirm;
                const confirmTitle = link.dataset.confirmTitle;
                
                const confirmed = await window.customConfirm(confirmMessage, confirmTitle);
                
                if (confirmed) {
                    window.location.href = link.href;
                }
            }
        });
        
        // Handle buttons with data-confirm attribute
        document.addEventListener('click', async (e) => {
            const button = e.target.closest('button[data-confirm]:not([type="submit"])');
            
            if (button) {
                e.preventDefault();
                
                const confirmMessage = button.dataset.confirm;
                const confirmTitle = button.dataset.confirmTitle;
                
                const confirmed = await window.customConfirm(confirmMessage, confirmTitle);
                
                if (confirmed && button.onclick) {
                    button.onclick();
                }
            }
        });
    });
    
})();