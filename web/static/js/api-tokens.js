/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

// API Tokens Management JavaScript (Minimal - Server-side rendered)
(function() {
    'use strict';

    let tokenToDelete = null;

    // Toggle expiration field
    const hasExpirationCheckbox = document.getElementById('hasExpiration');
    if (hasExpirationCheckbox) {
        hasExpirationCheckbox.addEventListener('change', function() {
            const expirationField = document.getElementById('expirationField');
            if (this.checked) {
                expirationField.classList.remove('hidden');
            } else {
                expirationField.classList.add('hidden');
                document.getElementById('expiresAt').value = '';
            }
        });
    }

    // Create token form submission (AJAX to avoid page reload)
    const createTokenForm = document.getElementById('createTokenForm');
    if (createTokenForm) {
        createTokenForm.addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const name = document.getElementById('tokenName').value;
            const hasExpiration = document.getElementById('hasExpiration').checked;
            const expiresAt = hasExpiration ? document.getElementById('expiresAt').value : null;

            const payload = {
                name: name,
                expires_at: expiresAt ? new Date(expiresAt).toISOString() : null
            };

            try {
                const response = await fetch('/api/v1/tokens', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });

                if (!response.ok) {
                    throw new Error('Failed to create token');
                }

                // Reload page to show new token
                // setTimeout(() => window.location.reload(), 200);

                const tokenData = await response.json();
            
                // Close create modal
                closeCreateTokenModal();

                displayNewToken(tokenData);

            } catch (error) {
                console.error('Error creating token:', error);
                showNotification('Failed to create API token', 'error');
            }
        });
    }

    // Display newly created token (ONE TIME)
    function displayNewToken(tokenData) {
        document.getElementById('displayedToken').value = tokenData.token;
        document.getElementById('displayTokenName').textContent = tokenData.name;
        document.getElementById('displayTokenCreated').textContent = formatDate(tokenData.created_at);
        document.getElementById('displayTokenExpires').textContent = 
            tokenData.expires_at ? formatDate(tokenData.expires_at) : 'Never';

        document.getElementById('displayTokenModal').showModal();
    }

    // Close Token Display and refresh
    window.closeDisplayTokenModal = function() {
        document.getElementById('displayTokenModal').close();
        setTimeout(() => window.location.reload(), 100);
    };


    // Copy token to clipboard
    window.copyToken = function() {
        const tokenInput = document.getElementById('displayedToken');
        tokenInput.select();
        tokenInput.setSelectionRange(0, 99999); // For mobile
        
        try {
            document.execCommand('copy');
            showNotification('Token copied to clipboard!', 'success');
        } catch (err) {
            console.error('Failed to copy:', err);
            showNotification('Failed to copy token', 'error');
        }
    };

    // Revoke token
    window.revokeToken = async function(tokenId, tokenName) {
        if (!confirm(`Revoke token "${tokenName}"? Applications using it will lose access immediately.`)) {
            return;
        }

        try {
            const response = await fetch(`/api/v1/tokens/${tokenId}/revoke`, {
                method: 'POST'
            });

            if (!response.ok) {
                throw new Error('Failed to revoke token');
            }

            showNotification('Token revoked successfully', 'success');
            
            // Reload page to update list
            setTimeout(() => window.location.reload(), 500);
            
        } catch (error) {
            console.error('Error revoking token:', error);
            showNotification('Failed to revoke token', 'error');
        }
    };

    // Delete token (show confirmation modal)
    window.deleteToken = function(tokenId, tokenName) {
        tokenToDelete = tokenId;
        document.getElementById('deleteTokenName').textContent = tokenName;
        document.getElementById('confirmDeleteModal').showModal();
    };

    // Confirm delete
    window.confirmDelete = async function() {
        if (!tokenToDelete) return;

        try {
            const response = await fetch(`/api/v1/tokens/${tokenToDelete}`, {
                method: 'DELETE'
            });

            if (!response.ok) {
                throw new Error('Failed to delete token');
            }

            showNotification('Token deleted successfully', 'success');
            closeConfirmDeleteModal();
            
            // Reload page to update list
            setTimeout(() => window.location.reload(), 500);
            
        } catch (error) {
            console.error('Error deleting token:', error);
            showNotification('Failed to delete token', 'error');
        }
    };

    // Modal helpers
    window.showCreateTokenModal = function() {
        const form = document.getElementById('createTokenForm');
        if (form) form.reset();
        
        const expirationField = document.getElementById('expirationField');
        if (expirationField) expirationField.classList.add('hidden');
        
        document.getElementById('createTokenModal').showModal();
    };

    window.closeCreateTokenModal = function() {
        document.getElementById('createTokenModal').close();
    };

    window.closeConfirmDeleteModal = function() {
        tokenToDelete = null;
        document.getElementById('confirmDeleteModal').close();
    };

    console.log('API Tokens page initialized');

})();
