/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

// KindredCard - Settings Page JavaScript
(function() {
    'use strict';

    console.log('Settings page initializing...');

    // Tab switching functionality
    const tabs = document.querySelectorAll('.tab');
    const tabContents = document.querySelectorAll('.tab-content');
    
    if (tabs.length > 0) {
        tabs.forEach(tab => {
            tab.addEventListener('click', function(e) {
                e.preventDefault();
                
                console.log('Tab clicked:', this.dataset.tab);
                
                // Remove active class from all tabs
                tabs.forEach(t => t.classList.remove('tab-active'));
                
                // Hide all tab contents
                tabContents.forEach(c => {
                    c.classList.add('hidden');
                    c.style.display = 'none';
                });
                
                // Activate clicked tab
                this.classList.add('tab-active');
                
                // Show corresponding content
                const tabName = this.dataset.tab;
                const tabContent = document.getElementById(tabName + 'Tab');
                
                if (tabContent) {
                    tabContent.classList.remove('hidden');
                    tabContent.style.display = 'block';
                    console.log('Tab content shown:', tabName);
                } else {
                    console.error('Tab content not found:', tabName + 'Tab');
                }
                
                // Save to URL hash for persistence across page reloads
                window.location.hash = tabName;
            });
        });

        // Restore active tab from URL hash on page load
        function restoreActiveTab() {
            const hash = window.location.hash.substring(1); // Remove '#' from hash
            
            if (hash) {
                console.log('Restoring tab from URL hash:', hash);
                
                // Find the tab with matching data-tab
                const tab = document.querySelector(`.tab[data-tab="${hash}"]`);
                if (tab) {
                    // Trigger click on that tab to activate it
                    tab.click();
                    return;
                }
            }
            
            // If no hash or invalid hash, show first tab
            console.log('No valid hash, showing first tab');
            const firstTab = document.querySelector('.tab-active');
            if (firstTab) {
                const tabName = firstTab.dataset.tab;
                const firstContent = document.getElementById(tabName + 'Tab');
                if (firstContent) {
                    firstContent.classList.remove('hidden');
                    firstContent.style.display = 'block';
                    console.log('First tab content shown:', tabName);
                }
            }
        }

        // Restore tab on page load
        restoreActiveTab();
        
    } else {
        console.error('No tabs found! Check HTML structure.');
    }

    // ========================================
    // CONTACTS TAB FUNCTIONS
    // ========================================

    // Import vCard
    window.importVCard = async function() {
        const fileInput = document.getElementById('vcardImport');
        const file = fileInput.files[0];
        
        if (!file) {
            showNotification('Please select a file', 'warning');
            return;
        }

        const formData = new FormData();
        formData.append('vcard', file);

        try {
            const response = await fetch('/api/v1/contacts/import', {
                method: 'POST',
                body: formData
            });

            if (response.ok) {
                const result = await response.json();
                showNotification(`Imported ${result.count} contact(s)`, 'success');
                fileInput.value = '';
                setTimeout(() => location.reload(), 1500);
            } else {
                throw new Error('Import failed');
            }
        } catch (error) {
            console.error('Import error:', error);
            showNotification('Failed to import contacts', 'error');
        }
    };

    // Export all contacts
    window.exportAllContacts = async function(asJSON) {
        try {
            // 1. Determine configuration based on format
            const endpoint = asJSON ? '/api/v1/contacts/export/json' : '/api/v1/contacts/export/vcard';
            const extension = asJSON ? 'json' : 'vcf';
            const date = new Date().toISOString().split('T')[0];

            // 2. Single fetch call
            const response = await fetch(endpoint);
            if (!response.ok) throw new Error('Export failed');

            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);

            // 3. Setup download
            const a = document.createElement('a');
            a.href = url;
            a.download = `kindredcard-contacts-${date}.${extension}`;
            
            document.body.appendChild(a);
            a.click();
            
            // Cleanup
            window.URL.revokeObjectURL(url);
            a.remove();
            showNotification('Contacts exported successfully', 'success');
        } catch (error) {
            console.error('Export error:', error);
            showNotification('Failed to export contacts', 'error');
        }
    };

    // Delete all contacts
    window.deleteAllContacts = async function() {
        const confirmed = confirm('⚠️ DELETE ALL CONTACTS?\n\nThis will permanently delete ALL contacts and cannot be undone.\n\nType "DELETE ALL" in the next prompt to confirm.');
        
        if (!confirmed) return;

        const verification = prompt('Type "DELETE ALL" (in capital letters) to confirm:');
        
        if (verification !== 'DELETE ALL') {
            showNotification('Verification failed - contacts not deleted', 'info');
            return;
        }

        try {
            const response = await fetch('/api/v1/contacts', {
                method: 'DELETE',
                headers: { 'Content-Type': 'application/json' }
            });

            if (response.ok) {
                showNotification('All contacts deleted', 'success');
                setTimeout(() => location.reload(), 1500);
            } else {
                throw new Error('Delete failed');
            }
        } catch (error) {
            console.error('Delete error:', error);
            showNotification('Failed to delete contacts', 'error');
        }
    };

    // Find duplicates
    window.findDuplicates = async function() {
        showNotification('Scanning for duplicates...', 'info');
        
        try {
            const response = await fetch('/api/v1/contacts/duplicates');
            
            if (response.ok) {
                const result = await response.json();
                if (result.duplicates && result.duplicates.length > 0) {
                    showNotification(`Found ${result.duplicates.length} potential duplicate(s)`, 'warning');
                    // TODO: Show duplicates modal
                } else {
                    showNotification('No duplicates found!', 'success');
                }
            } else {
                throw new Error('Scan failed');
            }
        } catch (error) {
            console.error('Duplicate scan error:', error);
            showNotification('Failed to scan for duplicates', 'error');
        }
    };

    // ========================================
    // NOTIFICATIONS TAB FUNCTIONS
    // ========================================

    // Open add notification modal
    window.openAddWebhookModal = function() {
        const modal = document.getElementById('notificationModal');
        const form = document.getElementById('notificationForm');
        
        if (form) {
            form.reset();
            document.getElementById('notification_id').value = '';
            document.getElementById('notification_provider_type').value = "discord";
            document.getElementById('notification_enabled').checked = true;
            document.getElementById('notification_include_birthdays').checked = true;
            document.getElementById('notification_include_anniversaries').checked = true;

            //show the webhook_url, hide target_address
            document.getElementById('notification_webhook_url').classList.remove('hidden');
            document.getElementById('notification_target_address').classList.add('hidden');

            //update requirements
            document.querySelector('input[name="notification_webhook_url"]').required = true;
            document.querySelector('input[name="notification_target_address"]').required = false;
        }

        // Update title
        document.getElementById('notificationModalTitle').textContent = 'Add Discord Webhook';
        document.getElementById('notificationSubmitText').textContent = 'Save Webhook';
        
        if (modal) modal.showModal();
    };

    // Open add notification modal
    window.openAddEmailModal = function() {
        const modal = document.getElementById('notificationModal');
        const form = document.getElementById('notificationForm');
        
        if (form) {
            form.reset();
            document.getElementById('notification_id').value = '';
            document.getElementById('notification_provider_type').value = "smtp";
            document.getElementById('notification_enabled').checked = true;
            document.getElementById('notification_include_birthdays').checked = true;
            document.getElementById('notification_include_anniversaries').checked = true;

            //hide the webhook_url, show target_address
            document.getElementById('notification_webhook_url').classList.add('hidden');
            document.getElementById('notification_target_address').classList.remove('hidden');

            //update requirements
            document.querySelector('input[name="notification_webhook_url"]').required = false;
            document.querySelector('input[name="notification_target_address"]').required = true;
        }

        // Update title
        document.getElementById('notificationModalTitle').textContent = 'Add Email Notification';
        document.getElementById('notificationSubmitText').textContent = 'Save Notification';
        
        if (modal) modal.showModal();
    };

    // Close notification modal
    window.closeNotificationModal = function() {
        const modal = document.getElementById('notificationModal');
        if (modal) modal.close();
    };

    // Edit notification
    window.editNotification = async function(id) {
        try {
            const response = await fetch(`/api/v1/notification-settings/${id}`);
            if (!response.ok) throw new Error('Failed to load notification');

            const webhookUrlField = document.getElementById('notification_webhook_url')
            const targetAddressField = document.getElementById('notification_target_address')

            const notification = await response.json();
            
            // Populate form
            document.getElementById('notification_id').value = notification.id;
            document.getElementById('notification_provider_type').value = notification.provider_type ?? "discord";
            document.getElementById('notification_name').value = notification.name;
            webhookUrlField.querySelector('input').value = notification.webhook_url || "";
            targetAddressField.querySelector('input').value = notification.target_address || "";
            document.getElementById('notification_days_look_ahead').value = notification.days_look_ahead ?? 0;
            document.getElementById('notification_notification_time').value = notification.notification_time ?? "09:00";
            document.getElementById('notification_include_birthdays').checked = notification.include_birthdays || false;
            document.getElementById('notification_include_anniversaries').checked = notification.include_anniversaries || false;
            document.getElementById('notification_include_event_dates').checked = notification.include_event_dates || false;
            document.getElementById('notification_other_event_regex').value = notification.other_event_regex ?? '';
            document.getElementById('notification_enabled').checked = notification.enabled || false;
            
            // Update modal based on type
            let modalTitle;
            let submitText;
            if ( notification.provider_type == "smtp" ) {
                modalTitle = 'Edit Email Target';
                submitText = 'Update Notification';
                webhookUrlField.classList.add('hidden');
                targetAddressField.classList.remove('hidden');
            } else {
                modalTitle = 'Edit Discord Webhook';
                submitText = 'Update Webhook';
                webhookUrlField.classList.remove('hidden');
                targetAddressField.classList.add('hidden');
            }

            document.getElementById('notificationModalTitle').textContent = modalTitle;
            document.getElementById('notificationSubmitText').textContent = submitText;
            
            // Open modal
            document.getElementById('notificationModal').showModal();
        } catch (error) {
            console.error('Load notification error:', error);
            showNotification('Failed to load notification settings', 'error');
        }
    };

    // Test notification
    window.testNotification = async function(id) {
        const button = event.target.closest('button');
        button.classList.add('loading');
        button.disabled = true;

        try {
            const response = await fetch(`/api/v1/notification-settings/${id}/test`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' }
            });

            if (response.ok) {
                showNotification('Test notification sent!', 'success');
            } else {
                const error = await response.json();
                throw new Error(error.message || 'Test failed');
            }
        } catch (error) {
            console.error('Test notification error:', error);
            showNotification(`Failed to send test: ${error.message}`, 'error');
        } finally {
            button.classList.remove('loading');
            button.disabled = false;
        }
    };

    // Delete notification
    window.deleteNotification = async function(id) {
        const confirmed = await customConfirm('Are you sure you want to delete this notification?')
        if (!confirmed) return;

        try {
            const response = await fetch(`/api/v1/notification-settings/${id}`, {
                method: 'DELETE',
                headers: { 'Content-Type': 'application/json' }
            });

            if (response.ok) {
                showNotification('Notification deleted', 'success');
                setTimeout(() => location.reload(), 800);
            } else {
                throw new Error('Delete failed');
            }
        } catch (error) {
            console.error('Delete notification error:', error);
            showNotification('Failed to delete notification', 'error');
        }
    };

    // Notification form submission
    notificationForm.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        // 1. Use FormData to get all values at once
        const formData = new FormData(notificationForm);
        const notificationId = document.getElementById('notification_id').value;
        const isEdit = !!notificationId;

        // 2. Build the data object safely
        const data = {
            name: formData.get('notification_name'),
            provider_type: formData.get('notification_provider_type'),
            webhook_url: formData.get('notification_webhook_url') || "",
            target_address: formData.get('notification_target_address') || "",
            days_look_ahead: parseInt(formData.get('notification_days_look_ahead')) || 0,
            notification_time: formData.get('notification_notification_time'),
            include_birthdays: formData.get('notification_include_birthdays') === 'on' || document.getElementById('notification_include_birthdays').checked,
            include_anniversaries: formData.get('notification_include_anniversaries') === 'on' || document.getElementById('notification_include_anniversaries').checked,
            include_event_dates: formData.get('notification_include_event_dates') === 'on' || document.getElementById('notification_include_event_dates').checked,
            other_event_regex: formData.get('notification_other_event_regex') || "",
            enabled: formData.get('notification_enabled') === 'on' || document.getElementById('notification_enabled').checked,
        };

        // Validation
        if (data.provider_type === "discord" && (!data.webhook_url || !data.webhook_url.startsWith('https://discord.com/api/webhooks/'))) {
            showNotification('Invalid Discord webhook URL', 'error');
            return;
        }
        
        if (data.provider_type === "smtp" && !data.target_address.includes('@')) {
            showNotification('Invalid email address', 'error');
            return;
        }

        try {
            const url = isEdit 
                ? `/api/v1/notification-settings/${notificationId}`
                : '/api/v1/notification-settings';

            const response = await fetch(url, {
                method: isEdit ? 'PUT' : 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });

            if (response.ok) {
                showNotification(`Notification ${isEdit ? 'updated' : 'created'} successfully`, 'success');
                // Assuming this function exists to hide the daisyUI modal
                if (window.notification_modal) window.notification_modal.close(); 
                setTimeout(() => location.reload(), 800);
            } else {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Save failed');
            }
        } catch (error) {
            console.error('Save notification error:', error);
            showNotification(error.message, 'error');
        }
    });

    // ========================================
    // SECURITY TAB FUNCTIONS
    // ========================================

    // Timezone things
    window.savePreferences = async function(prefs) {
        try {
            const response = await fetch('/api/v1/user/preferences', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(prefs)
            });
            if (response.ok) {
                console.log("Preferences saved successfully");
            }
        } catch (err) {
            console.error("Failed to save preferences:", err);
        }
    }

    window.updateTimezone = function(tz) {
        savePreferences({ 
            theme: document.documentElement.getAttribute('data-theme'),
            timezone: tz 
        });
    };

    window.detectTimezone = function() {
        const detectedTZ = Intl.DateTimeFormat().resolvedOptions().timeZone;
        const select = document.getElementById('timezone-select');
        
        if (select) {
            select.value = detectedTZ;
            updateTimezone(detectedTZ);
            
            if (window.showNotification) {
                showNotification(`Detected: ${detectedTZ}`, 'success');
            }
        }
    };

    // Password change form validation
    const passwordForm = document.querySelector('form[action="/settings/password"]');
    if (passwordForm) {
        passwordForm.addEventListener('submit', function(e) {
            const currentPassword = this.querySelector('input[name="current_password"]').value;
            const newPassword = this.querySelector('input[name="new_password"]').value;
            const confirmPassword = this.querySelector('input[name="confirm_password"]').value;
            
            if (!currentPassword || !newPassword || !confirmPassword) {
                e.preventDefault();
                showNotification('Please fill in all password fields', 'error');
                return false;
            }
            
            if (newPassword !== confirmPassword) {
                e.preventDefault();
                showNotification('New passwords do not match', 'error');
                return false;
            }
            
            if (newPassword.length < 8) {
                e.preventDefault();
                showNotification('New password must be at least 8 characters', 'error');
                return false;
            }
            
            if (newPassword === currentPassword) {
                e.preventDefault();
                showNotification('New password must be different from current password', 'warning');
                return false;
            }
        });
    }

    // Delete account confirmation
    window.confirmDeleteAccount = function() {
        if (confirm('Are you ABSOLUTELY sure? This action cannot be undone. All your data will be permanently deleted.')) {
            const confirmation = prompt('Type "DELETE" in all caps to confirm:');
            if (confirmation === 'DELETE') {
                // Show loading state
                const deleteBtn = event.target;
                deleteBtn.classList.add('loading');
                deleteBtn.disabled = true;
                
                // Submit delete request
                fetch('/settings/delete-account', { 
                    method: 'DELETE',
                    headers: {
                        'Content-Type': 'application/json'
                    }
                })
                .then(response => {
                    if (response.ok) {
                        showNotification('Account deleted. Redirecting...', 'info');
                        setTimeout(() => window.location.href = '/logout', 1500);
                    } else {
                        throw new Error('Delete failed');
                    }
                })
                .catch(error => {
                    console.error('Delete error:', error);
                    showNotification('Failed to delete account', 'error');
                    deleteBtn.classList.remove('loading');
                    deleteBtn.disabled = false;
                });
            } else if (confirmation !== null) {
                showNotification('Confirmation text did not match', 'warning');
            }
        }
    };

    console.log('Settings page initialized successfully');
})();