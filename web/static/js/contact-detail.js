/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

// KindredCard - Contact Detail Page JavaScript
(function() {
    'use strict';

    let hasChanges = false;
    const saveButton = document.getElementById('saveButton');

    // Initialize change detection
    function initChangeDetection() {
        // Listen to all form inputs
        const form = document.getElementById('contactForm');
        if (!form) return;

        // Track changes on all inputs, selects, and textareas
        form.addEventListener('input', markAsChanged);
        form.addEventListener('change', markAsChanged);
    }

    // Mark form as changed and update save button
    function markAsChanged() {
        if (!hasChanges) {
            hasChanges = true;
            if (saveButton) {
                saveButton.classList.remove('btn-ghost');
                saveButton.classList.add('btn-accent');
                saveButton.disabled = false;
            }
        }
    }

    // Mark form as changed (can be called externally)
    window.markFormChanged = markAsChanged;

    // Show saving overlay
    function showSavingOverlay() {
        const overlay = document.createElement('div');
        overlay.id = 'savingOverlay';
        overlay.className = 'saving-overlay';
        overlay.innerHTML = `
            <div class="saving-content">
                <div class="saving-spinner" style="animation: spin 0.8s linear infinite;"></div>
                <div class="saving-text">Saving...</div>
            </div>
        `;
        document.body.appendChild(overlay);
        
        // Add keyframes if not already present
        if (!document.getElementById('spin-keyframes')) {
            const style = document.createElement('style');
            style.id = 'spin-keyframes';
            style.textContent = `
                @keyframes spin {
                    0% { transform: rotate(0deg); }
                    100% { transform: rotate(360deg); }
                }
            `;
            document.head.appendChild(style);
        }
        
        // Trigger animation
        requestAnimationFrame(() => {
            overlay.classList.add('active');
        });
    }

    // Hide saving overlay
    function hideSavingOverlay() {
        const overlay = document.getElementById('savingOverlay');
        if (overlay) {
            overlay.classList.remove('active');
            setTimeout(() => overlay.remove(), 300);
        }
    }

    // Initialize on page load
    initChangeDetection();

    let emailCount = document.querySelectorAll('#emailsList > div').length;
    let phoneCount = document.querySelectorAll('#phonesList > div').length;
    let addressCount = document.querySelectorAll('#addressesList > .card').length;
    let orgCount = document.querySelectorAll('#orgsList > .card').length;
    let urlCount = document.querySelectorAll('#urlsList > div').length;
    let otherDatesCount = document.querySelectorAll('#otherDatesList > div').length;

    // Add email field
    window.addEmail = function() {
        const container = document.getElementById('emailsList');
        const index = emailCount++;
        
        const div = document.createElement('div');
        div.className = 'flex gap-2 items-start';
        div.innerHTML = `
            <input type="email" name="emails[${index}][email]" placeholder="email@example.com" class="input input-bordered flex-1" required>
            <select name="emails[${index}][type]" class="select select-bordered">
                <option value="home">Home</option>
                <option value="work">Work</option>
                <option value="other">Other</option>
            </select>
            <label class="label cursor-pointer gap-2">
                <input type="checkbox" name="emails[${index}][is_primary]" class="checkbox checkbox-primary">
                <span class="label-text">Primary</span>
            </label>
            <button type="button" class="btn btn-ghost btn-sm btn-square" onclick="this.parentElement.remove(); markFormChanged();">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        `;
        
        container.appendChild(div);
        markFormChanged();
    };

    // Add phone field
    window.addPhone = function() {
        const container = document.getElementById('phonesList');
        const index = phoneCount++;
        
        const div = document.createElement('div');
        div.className = 'flex gap-2 items-start';
        div.innerHTML = `
            <input type="tel" name="phones[${index}][phone]" placeholder="(555) 123-4567" class="input input-bordered flex-1" required>
            <select name="phones[${index}][type]" class="select select-bordered">
                <option value="cell">Cell</option>
                <option value="home">Home</option>
                <option value="work">Work</option>
                <option value="other">Other</option>
            </select>
            <label class="label cursor-pointer gap-2">
                <input type="checkbox" name="phones[${index}][is_primary]" class="checkbox checkbox-primary">
                <span class="label-text">Primary</span>
            </label>
            <button type="button" class="btn btn-ghost btn-sm btn-square" onclick="this.parentElement.remove(); markFormChanged();">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        `;
        
        container.appendChild(div);
        markFormChanged();
    };

    // Add address field
    window.addAddress = function() {
        const container = document.getElementById('addressesList');
        const index = addressCount++;
        
        const div = document.createElement('div');
        div.className = 'card bg-base-200 shadow-md p-4';
        div.innerHTML = `
            <div class="flex justify-between items-start mb-3">
                <select name="addresses[${index}][type]" class="select select-bordered select-sm">
                    <option value="home">Home</option>
                    <option value="work">Work</option>
                    <option value="other">Other</option>
                </select>
                <div class="flex gap-2">
                    <label class="label cursor-pointer gap-2">
                        <input type="checkbox" name="addresses[${index}][is_primary]" class="checkbox checkbox-primary checkbox-sm">
                        <span class="label-text">Primary</span>
                    </label>
                    <button type="button" class="btn btn-ghost btn-sm btn-square" onclick="this.closest('.card').remove()">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>
            </div>
            <div class="space-y-2">
                <div class="grid grid-cols-5 gap-2">
                    <input type="text" name="addresses[${index}][street]" placeholder="Street Address" class="input input-bordered input-sm col-span-3">
                    <input type="text" name="addresses[${index}][extended_street]" placeholder="Apt/Ste" class="input input-bordered input-sm col-span-2">
                </div>

                <div class="grid grid-cols-5 gap-2">
                    <input type="text" name="addresses[${index}][city]" placeholder="City" class="input input-bordered input-sm col-span-3">
                    <input type="text" name="addresses[${index}][state]" placeholder="State" class="input input-bordered input-sm col-span-2">
                </div>

                <div class="grid grid-cols-2 gap-2">
                    <input type="text" name="addresses[${index}][postal_code]" placeholder="ZIP" class="input input-bordered input-sm">
                    <input type="text" name="addresses[${index}][country]" placeholder="Country" class="input input-bordered input-sm">
                </div>
            </div>
        `;
        
        container.appendChild(div);
        markFormChanged();
    };

    // Add organization field
    window.addOrganization = function() {
        const container = document.getElementById('orgsList');
        const index = orgCount++;
        
        const div = document.createElement('div');
        div.className = 'card bg-base-200 shadow-md p-4';
        div.innerHTML = `
            <div class="flex justify-end mb-2">
                <button type="button" class="btn btn-ghost btn-sm btn-square" onclick="this.closest('.card').remove()">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                <input type="text" name="organizations[${index}][name]" placeholder="Company Name" class="input input-bordered input-sm">
                <input type="text" name="organizations[${index}][title]" placeholder="Job Title" class="input input-bordered input-sm">
                <input type="text" name="organizations[${index}][department]" placeholder="Department" class="input input-bordered input-sm col-span-2">
            </div>
        `;
        
        container.appendChild(div);
        markFormChanged();
    };

    // Add URL field
    window.addURL = function() {
        const container = document.getElementById('urlsList');
        const index = urlCount++;
        
        const div = document.createElement('div');
        div.className = 'flex gap-2 items-center';
        div.innerHTML = `
            <input type="url" name="urls[${index}][url]" placeholder="https://example.com" class="input input-bordered input-sm flex-1">
            <select name="urls[${index}][type]" class="select select-bordered select-sm">
                <option value="website">Website</option>
                <option value="social">Social</option>
                <option value="other">Other</option>
            </select>
            <button type="button" class="btn btn-ghost btn-sm btn-square" onclick="this.parentElement.remove(); markFormChanged();">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        `;
        
        container.appendChild(div);
        markFormChanged();
    };

    // Add other date function
    window.addOtherDate = function() {
        const container = document.getElementById('otherDatesList');
        const index = otherDatesCount++;
        
        const div = document.createElement('div');
        div.className = 'flex gap-2 items-start border-l-4 border-primary pl-3';
        div.innerHTML = `
            <div class="flex-1 space-y-2">
                <input type="text" name="other_dates[${index}][event_name]" placeholder="Event name" class="input input-bordered input-sm w-full" required>
                <div class="grid gap-2" style="grid-template-columns: 1fr 0.8fr 0.8fr;">
                    <select name="other_dates[${index}][event_date_month]" class="select select-bordered select-sm" data-date-group="other_dates_${index}">
                        <option value="">*Month</option>
                        <option value="1">Jan</option>
                        <option value="2">Feb</option>
                        <option value="3">Mar</option>
                        <option value="4">Apr</option>
                        <option value="5">May</option>
                        <option value="6">Jun</option>
                        <option value="7">Jul</option>
                        <option value="8">Aug</option>
                        <option value="9">Sep</option>
                        <option value="10">Oct</option>
                        <option value="11">Nov</option>
                        <option value="12">Dec</option>
                    </select>
                    <select name="other_dates[${index}][event_date_day]" class="select select-bordered select-sm" data-date-group="other_dates_${index}">
                        <option value="">*Day</option>
                        ${Array.from({length: 31}, (_, i) => `<option value="${i+1}">${i+1}</option>`).join('')}
                    </select>
                    <input type="number" name="other_dates[${index}][event_date_year]" placeholder="Year" min="1900" max="2100" class="input input-bordered input-sm">
                </div>
            </div>
            <button type="button" class="btn btn-ghost btn-xs btn-circle" onclick="this.parentElement.remove(); markFormChanged();" title="Delete event">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        `;
        
        container.appendChild(div);
        
        // Add validation listeners to new selects
        div.querySelectorAll('select[data-date-group]').forEach(select => {
            select.addEventListener('change', function() {
                const dateGroup = this.dataset.dateGroup;
                validateDateGroup(dateGroup);
            });
        });
        
        markFormChanged();
    };

    // Avatar preview
    window.openAvatarModal = function() {
        const avatarImage = document.getElementById('avatarPreview');
        const modalImage = document.getElementById('avatarModalImage');
        
        if (avatarImage && modalImage) {
            modalImage.src = avatarImage.src;
            openModal('avatarModal');
        }
    };

    // Avatar upload
    window.uploadAvatar = async function() {
        const fileInput = document.getElementById('avatarInput');
        const file = fileInput.files[0];
        
        if (!file) return;

        // Validate file type
        if (!file.type.startsWith('image/')) {
            showNotification('Please select an image file', 'error');
            return;
        }

        // Validate file size (max 2MB)
        if (file.size > 2 * 1024 * 1024) {
            showNotification('Image must be less than 2MB', 'error');
            return;
        }

        // Convert to base64
        const reader = new FileReader();
        reader.onload = async function(e) {
            const base64 = e.target.result.split(',')[1]; // Remove data:image/jpeg;base64, prefix

            const contactId = window.location.pathname.split('/').pop();
            
            try {
                const response = await fetch(`/api/v1/contacts/${contactId}/avatar`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ avatar: base64 })
                });
                
                if (response.ok) {
                    showNotification('Avatar updated', 'success');
                    setTimeout(() => window.location.reload(), 500);
                } else {
                    throw new Error('Upload failed');
                }
            } catch (error) {
                console.error('Avatar upload error:', error);
                showNotification('Failed to upload avatar', 'error');
            }
        };
        reader.readAsDataURL(file);
    };

    // Delete avatar
    window.deleteAvatar = async function() {
        if (!confirm('Are you sure you want to remove the profile photo?')) {
            return;
        }

        const contactId = window.location.pathname.split('/').pop();
        
        try {
            const response = await fetch(`/api/v1/contacts/${contactId}/avatar`, {
                method: 'DELETE'
            });

            if (response.ok) {
                showNotification('Photo removed', 'success');
                setTimeout(() => window.location.reload(), 500);
            } else {
                throw new Error('Failed to delete avatar');
            }
        } catch (error) {
            console.error('Error:', error);
            showNotification('Failed to remove photo', 'error');
        }
    };

    // Export vCard
    window.exportVCard = async function(fullName) {
        const contactId = window.location.pathname.split('/').pop();
        
        try {
            const response = await fetch(`/api/v1/contacts/${contactId}/vcard`);
            if (!response.ok) throw new Error('Export failed');
            
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `${fullName}.vcf`;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            a.remove();
            
            showNotification('Contact exported', 'success');
        } catch (error) {
            console.error('Export error:', error);
            showNotification('Failed to export contact', 'error');
        }
    };

    // Delete contact with modal
    window.openDeleteModal = function() {
        const modal = document.getElementById('deleteModal');
        if (modal) {
            modal.showModal();
        }
    };

    window.closeDeleteModal = function() {
        const modal = document.getElementById('deleteModal');
        if (modal) {
            modal.close();
        }
    };

    window.confirmDelete = async function() {
        const contactId = window.location.pathname.split('/').pop();
        
        try {
            const response = await fetch(`/api/v1/contacts/${contactId}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                showNotification('Contact deleted successfully', 'success');
                closeDeleteModal();
                setTimeout(() => window.location.href = '/', 800);
            } else {
                throw new Error('Delete failed');
            }
        } catch (error) {
            console.error('Delete error:', error);
            showNotification('Failed to delete contact', 'error');
        }
    };

    // Remove relationship
    window.removeRelationship = async function(relationshipId) {
        if (!confirm('Remove this relationship?')) {
            return;
        }

        try {
            const response = await fetch(`/api/v1/relationships/${relationshipId}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                showNotification('Relationship removed', 'success');
                setTimeout(() => location.reload(), 500);
            } else {
                throw new Error('Delete failed');
            }
        } catch (error) {
            console.error('Error:', error);
            showNotification('Failed to remove relationship', 'error');
        }
    };

    // Remove relationship
    window.removeOtherRelationship = async function(otherRelationshipId) {
        if (!confirm('Remove this other relationship?')) {
            return;
        }

        try {
            const response = await fetch(`/api/v1/other-relationships/${otherRelationshipId}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                showNotification('Other Relationship removed', 'success');
                setTimeout(() => location.reload(), 500);
            } else {
                throw new Error('Delete failed');
            }
        } catch (error) {
            console.error('Error:', error);
            showNotification('Failed to remove other relationship', 'error');
        }
    };

    // Open add relationship modal
    window.openAddRelationshipModal = function() {
        const modal = document.getElementById('addRelationshipModal');
        if (modal) {
            // Reset form
            document.getElementById('relatedContactSelect').value = '';
            document.getElementById('relationshipTypeSelect').value = '';
            document.getElementById('newContactFormSection').classList.add('hidden');
            
            // Clear new contact form
            document.getElementById('new_given_name').value = '';
            document.getElementById('new_middle_name').value = '';
            document.getElementById('new_family_name').value = '';
            document.getElementById('new_gender').value = '';
            
            modal.showModal();
        }
    };

    // Close relationship modal
    window.closeRelationshipModal = function() {
        const modal = document.getElementById('addRelationshipModal');
        if (modal) {
            modal.close();
        }
    };

    // Handle contact selection change
    window.handleRelatedContactChange = function() {
        const select = document.getElementById('relatedContactSelect');
        const newContactForm = document.getElementById('newContactFormSection');
        
        if (select.value === 'new') {
            newContactForm.classList.remove('hidden');
        } else {
            newContactForm.classList.add('hidden');
        }
    };

    // Save relationship
    window.saveRelationship = async function() {
        const contactId = window.location.pathname.split('/').pop();
        const relatedContactSelect = document.getElementById('relatedContactSelect');
        const relationshipTypeSelect = document.getElementById('relationshipTypeSelect');
        
        let relatedContactId = relatedContactSelect.value;

        // Validation
        if (!relatedContactId) {
            showNotification('Please select a contact', 'error');
            return;
        }

        if (!relationshipTypeSelect.value) {
            showNotification('Please select a relationship type', 'error');
            return;
        }

        // If creating new contact
        if (relatedContactId === 'new') {
            const givenName = document.getElementById('new_given_name').value.trim();
            
            if (!givenName) {
                showNotification('Please enter at least a first name', 'error');
                return;
            }
            
            try {
                const response = await fetch('/api/v1/contacts', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        given_name: givenName,
                        middle_name: document.getElementById('new_middle_name').value.trim(),
                        family_name: document.getElementById('new_family_name').value.trim(),
                        gender: document.getElementById('new_gender').value
                    })
                });
                
                if (response.ok) {
                    const newContact = await response.json();
                    relatedContactId = newContact.id;
                    showNotification('New contact created', 'success');
                } else {
                    const error = await response.json();
                    throw new Error(error.message || 'Failed to create contact');
                }
            } catch (error) {
                console.error('Error creating contact:', error);
                showNotification(`Failed to create new contact: ${error.message}`, 'error');
                return;
            }
        }
        
        const relationshipTypeId = relationshipTypeSelect.value;
        
        // Add relationship
        try {
            const response = await fetch(`/api/v1/contacts/${contactId}/relationships`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    related_contact_id: parseInt(relatedContactId),
                    relationship_type_id: parseInt(relationshipTypeId)
                })
            });
            
            if (response.ok) {
                showNotification('Relationship added successfully', 'success');
                closeRelationshipModal();
                setTimeout(() => window.location.reload(), 500);
            } else {
                const error = await response.json();
                throw new Error(error.message || 'Failed to add relationship');
            }
        } catch (error) {
            console.error('Error adding relationship:', error);
            showNotification(`Failed to add relationship: ${error.message}`, 'error');
        }
    };

    // Form submission
    const form = document.getElementById('contactForm');
    if (form) {
        form.addEventListener('submit', async function(e) {
            e.preventDefault();

            showSavingOverlay();
            
            const formData = new FormData(this);
            const contactId = window.location.pathname.split('/').pop();
            
            // Validate dates before submission
            if (!validateDates()) {
                return;
            }
            
            // Convert FormData to JSON
            const contact = {};
            
            // Simple fields
            contact.prefix = formData.get('prefix') || '';
            contact.given_name = formData.get('given_name') || '';
            contact.middle_name = formData.get('middle_name') || '';
            contact.family_name = formData.get('family_name') || '';
            contact.suffix = formData.get('suffix') || '';
            contact.nickname = formData.get('nickname') || '';
            contact.notes = formData.get('notes') || '';
            contact.gender = formData.get('gender') || '';

            // Adjust checkbox
            contact.exclude_from_sync = formData.get('exclude_from_sync') || false;
            if (contact.exclude_from_sync == "on") {
                contact.exclude_from_sync = true;
            }
            
            // Parse dates
            const birthdayJson = parseDateFields('birthday', formData);
            if (birthdayJson) {
                if (birthdayJson.hasYear) {
                    contact.birthday = birthdayJson.output;
                } else {
                    contact.birthday_month = Number(birthdayJson.month);
                    contact.birthday_day = Number(birthdayJson.day);
                }
            }

            const anniversaryJson = parseDateFields('anniversary', formData);
            if (anniversaryJson) {
                if (anniversaryJson.hasYear) {
                    contact.anniversary = anniversaryJson.output;
                } else {
                    contact.anniversary_month = Number(anniversaryJson.month);
                    contact.anniversary_day = Number(anniversaryJson.day);
                }
            }

            // Collect arrays
            contact.emails = [];
            contact.phones = [];
            contact.addresses = [];
            contact.organizations = [];
            contact.urls = [];
            contact.other_dates = [];
            
            // Process form data
            const uniqueKeys = new Set(formData.keys());

            for (let key of uniqueKeys) {
                const match = key.match(/^(\w+)\[(\d+)\]\[(\w+)\]$/);
                if (!match) continue;

                const [, collection, index, field] = match;
                const idx = parseInt(index);

                if (!contact[collection]) contact[collection] = [];
                if (!contact[collection][idx]) contact[collection][idx] = {};

                let processedValue;

                if (field === 'is_primary') {
                    processedValue = (formData.get(key) === 'on');
                } else if (field === 'type') {
                    processedValue = formData.getAll(key).filter(v => v !== "" && v !== "other");
                } else {
                    processedValue = formData.get(key);
                }

                contact[collection][idx][field] = processedValue;
            }
            
            // Clean up arrays AND format 'type' for Go compatibility
            contact.emails = (contact.emails || [])
                .filter(e => e && e.email)
                .map(e => ({ ...e, type: Array.isArray(e.type) ? e.type : (e.type ? [e.type] : []) }));

            contact.phones = (contact.phones || [])
                .filter(p => p && p.phone)
                .map(p => ({ ...p, type: Array.isArray(p.type) ? p.type : (p.type ? [p.type] : []) }));

            contact.addresses = (contact.addresses || [])
                .filter(a => a && (a.street || a.city))
                .map(a => ({ ...a, type: Array.isArray(a.type) ? a.type : (a.type ? [a.type] : []) }));

            contact.urls = (contact.urls || [])
                .filter(u => u && u.url)
                .map(u => ({ ...u, type: Array.isArray(u.type) ? u.type : (u.type ? [u.type] : []) }));

            contact.organizations = contact.organizations.filter(o => o && o.name);
            
            // Process other_dates using existing parseDateFields pattern
            contact.other_dates = contact.other_dates
                .filter(d => d && d.event_name) // Must have event name
                .map(d => {
                    // Create a pseudo-FormData for parseDateFields
                    const pseudoFormData = {
                        get: (key) => {
                            if (key.endsWith('_month')) return d.event_date_month;
                            if (key.endsWith('_day')) return d.event_date_day;
                            if (key.endsWith('_year')) return d.event_date_year;
                            return null;
                        }
                    };
                    
                    // Parse date using existing function
                    const dateJson = parseDateFields('event_date', pseudoFormData);
                    
                    const result = {
                        event_name: d.event_name
                    };
                    
                    if (dateJson) {
                        if (dateJson.hasYear) {
                            result.event_date = dateJson.output;
                        } else {
                            result.event_date_month = Number(dateJson.month);
                            result.event_date_day = Number(dateJson.day);
                        }
                    }
                    
                    return result;
                })
                .filter(d => d.event_date || (d.event_date_month && d.event_date_day)); // Must have valid date
            
            try {
                const response = await fetch(`/api/v1/contacts/${contactId}?source=GUI`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(contact)
                });

                if (response.ok) {
                    //showNotification('Contact saved successfully', 'success');
                    setTimeout(() => location.reload(), 1000);
                } else {
                    hideSavingOverlay();
                    throw new Error('Save failed');
                }
            } catch (error) {
                console.error('Save error:', error);
                hideSavingOverlay();
                showNotification('Failed to save contact', 'error');
            }
        });
    }

    // Parse date fields (month, day, year) into ISO date string or null
    function parseDateFields(prefix, formData) {
        const month = formData.get(`${prefix}_month`);
        const day = formData.get(`${prefix}_day`);
        const year = formData.get(`${prefix}_year`);

        if (!month || !day) {
            return null;
        }

        const mm = String(month).padStart(2, "0");
        const dd = String(day).padStart(2, "0");

        const output = year
            ? `${year}-${mm}-${dd}`
            : `--${mm}-${dd}`;

        return {
            output,
            hasYear: Boolean(year),
            year: year || null,
            month: month,
            day: day
        };
    }

    // Validate date fields
    function validateDates() {
        let isValid = true;
        
        // Validate birthday
        isValid = validateDateGroup('birthday') && isValid;
        
        // Validate anniversary
        isValid = validateDateGroup('anniversary') && isValid;

        // Validate all other_dates entries
        const otherDateContainers = document.querySelectorAll('#otherDatesList > div');
        otherDateContainers.forEach((container, index) => {
            const nameInput = container.querySelector('input[name^="other_dates"][name$="[event_name]"]');
            const monthSelect = container.querySelector('select[name^="other_dates"][name$="[event_date_month]"]');
            const daySelect = container.querySelector('select[name^="other_dates"][name$="[event_date_day]"]');
            
            if (!nameInput || !monthSelect || !daySelect) return;
            
            const name = nameInput.value.trim();
            const month = monthSelect.value;
            const day = daySelect.value;
            
            // If name is provided, validate the date using existing function
            if (name) {
                // Get the date group identifier from data attribute
                const dateGroup = monthSelect.dataset.dateGroup;
                if (dateGroup && !validateDateGroup(dateGroup)) {
                    isValid = false;
                }
            }
            
            // If month or day is provided, name must be provided
            if ((month || day) && !name) {
                nameInput.classList.add('input-error');
                showNotification('Event name is required when date is provided', 'error');
                isValid = false;
            } else {
                nameInput.classList.remove('input-error');
            }
        });
        
        return isValid;
    }

    // Validate a date group (month/day must both be set or both be empty)
    function validateDateGroup(prefix) {
        const monthSelect = document.querySelector(`select[name="${prefix}_month"]`);
        const daySelect = document.querySelector(`select[name="${prefix}_day"]`);
        
        if (!monthSelect || !daySelect) {
            return true; // Fields not found, skip validation
        }
        
        const month = monthSelect.value;
        const day = daySelect.value;
        
        // Both empty is OK
        if (!month && !day) {
            clearDateError(prefix);
            return true;
        }
        
        // Both filled is OK
        if (month && day) {
            // Validate day is valid for month
            if (!isValidDayForMonth(parseInt(month), parseInt(day))) {
                showDateError(prefix, 'Invalid day for selected month');
                return false;
            }
            clearDateError(prefix);
            return true;
        }
        
        // One filled, one empty is NOT OK
        if (month && !day) {
            showDateError(prefix, 'Day is required when month is selected');
            return false;
        }
        
        if (!month && day) {
            showDateError(prefix, 'Month is required when day is selected');
            return false;
        }
        
        return true;
    }

    // Check if day is valid for given month
    function isValidDayForMonth(month, day) {
        const daysInMonth = [31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];
        
        if (month < 1 || month > 12) return false;
        if (day < 1 || day > daysInMonth[month - 1]) return false;
        
        return true;
    }

    // Show date validation error
    function showDateError(prefix, message) {
        const monthSelect = document.querySelector(`select[name="${prefix}_month"]`);
        const daySelect = document.querySelector(`select[name="${prefix}_day"]`);
        
        // Add error classes
        monthSelect?.classList.add('select-error');
        daySelect?.classList.add('select-error');
        
        // Show error message
        const container = monthSelect?.closest('.form-control');
        if (container) {
            // Remove existing error
            const existingError = container.querySelector('.date-error-message');
            if (existingError) {
                existingError.remove();
            }
            
            // Add new error
            const errorLabel = document.createElement('label');
            errorLabel.className = 'label date-error-message';
            errorLabel.innerHTML = `<span class="label-text-alt text-error">${message}</span>`;
            container.appendChild(errorLabel);
        }
        
        //showNotification(message, 'error');
    }

    // Clear date validation error
    function clearDateError(prefix) {
        const monthSelect = document.querySelector(`select[name="${prefix}_month"]`);
        const daySelect = document.querySelector(`select[name="${prefix}_day"]`);
        
        // Remove error classes
        monthSelect?.classList.remove('select-error');
        daySelect?.classList.remove('select-error');
        
        // Remove error message
        const container = monthSelect?.closest('.form-control');
        const errorMessage = container?.querySelector('.date-error-message');
        if (errorMessage) {
            errorMessage.remove();
        }
    }

    // Add real-time validation listeners
    document.querySelectorAll('select[data-date-group]').forEach(select => {
        select.addEventListener('change', function() {
            const dateGroup = this.dataset.dateGroup;
            validateDateGroup(dateGroup);
        });
    });

    // Clear date function (birthday or anniversary)
    window.clearDate = function(dateType) {
        // Reset month select
        const monthSelect = document.getElementById(`${dateType}_month`);
        if (monthSelect) monthSelect.value = '';
        
        // Reset day select
        const daySelect = document.getElementById(`${dateType}_day`);
        if (daySelect) daySelect.value = '';
        
        // Reset year input
        const yearInput = document.getElementById(`${dateType}_year`);
        if (yearInput) yearInput.value = '';
        
        // Clear any validation errors
        clearDateError(dateType);
        
        // Mark form as changed
        markFormChanged();
        
        //showNotification(`${dateType.charAt(0).toUpperCase() + dateType.slice(1)} cleared`, 'info');
    };

    console.log('Contact detail page initialized');
})();
