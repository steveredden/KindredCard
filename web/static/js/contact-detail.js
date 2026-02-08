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

    /////////////////////
    // Add header handling
    /////////////////////

    window.evaluateHeaderChange = function() {
        const form = document.getElementById('contactForm');
        const btn = document.getElementById('saveButton');
        const fields = ['prefix', 'given_name', 'middle_name', 'family_name', 'suffix', 'gender', 'nickname', 'maiden_name'];
        
        const isDirty = fields.some(field => {
            const input = form.querySelector(`[name="${field}"]`);
            if (!input) return false;
            
            const current = input.value.trim();
            const original = (form.getAttribute(`data-original-${field}`) || "").trim();
            return current !== original;
        });

        btn.disabled = !isDirty;
        // Add visual cue
        if (isDirty) btn.classList.add('btn-active', 'text-warning');
        else btn.classList.remove('btn-active', 'text-warning');
    };

    document.getElementById('contactForm').addEventListener('submit', async function(e) {
        e.preventDefault();
        const btn = document.getElementById('saveButton');
        const contactId = window.location.pathname.split('/').pop();
        
        btn.disabled = true;
        btn.innerHTML = '<span class="loading loading-spinner"></span>';

        const formData = new FormData(this);
        const payload = Object.fromEntries(formData.entries());
        delete payload.avatar;  //this is handled on its own, though happens to live within the form
        // also handle maiden_name removal if Male
        if (payload.gender === 'M') {
            payload.maiden_name = "";
        }

        try {
            const res = await fetch(`/api/v1/contacts/${contactId}`, {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });

            if (res.ok) {
                showNotification('Contact names updated', 'success');
                setTimeout(() => location.reload(), 500);
            } else {
                throw new Error();
            }
        } catch (err) {
            showNotification('Update failed', 'error');
        } finally {
            btn.innerHTML = 'Save Changes';
        }
    });

    /////////////////////
    // end header handling
    /////////////////////

    /////////////////////
    // Add email handling
    /////////////////////

    window.markEmailsChanged = function() {
        document.getElementById('saveEmailsBtn').disabled = false;
        document.getElementById('saveEmailsBtn').classList.add('btn-pulse'); // Optional flair
    };

    window.addEmail = function() {
        const container = document.getElementById('emailsList');
        const optionsHtml = document.getElementById('emailTypeOptionsTmpl').innerHTML;
        
        const div = document.createElement('div');
        div.className = 'flex gap-2 items-start email-row';

        div.innerHTML = `
            <input type="email" name="email" placeholder="email@example.com" class="input input-bordered flex-1" required>
            <select name="label_type_id" class="select select-bordered">
                ${optionsHtml}
            </select>
            <label class="label cursor-pointer gap-2">
                <input type="checkbox" name="is_primary" class="checkbox checkbox-primary">
                <span class="label-text">Primary</span>
            </label>
            <button type="button" class="btn btn-ghost btn-sm btn-square" onclick="removeEmailRow(this)">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        `;
        
        container.appendChild(div);
        markFormChanged();
    };

    let deletedEmailIds = [];

    window.removeEmailRow = function(btn) {
        const row = btn.closest('.email-row');
        const id = row.getAttribute('data-id');
        
        if (id) {
            deletedEmailIds.push(parseInt(id));
        }
        
        row.remove();
        markEmailsChanged();
    };

    window.saveEmailsOnly = async function() {
        const btn = document.getElementById('saveEmailsBtn');
        const contactId = window.location.pathname.split('/').pop();
        const emailRows = document.querySelectorAll('.email-row');
        
        btn.disabled = true;
        const requests = [];

        // 1. Handle Deletions first
        deletedEmailIds.forEach(id => {
            requests.push(fetch(`/api/v1/contacts/${contactId}/emails/${id}`, { method: 'DELETE' }));
        });

        // 2. Loop through rows to determine POST (new) vs PATCH (update)
        emailRows.forEach(row => {
            const id = row.getAttribute('data-id');

            const currentEmail = row.querySelector('[name="email"]').value;
            const currentType = parseInt(row.querySelector('[name="label_type_id"]').value);
            const currentPrimary = !!row.querySelector('[name="is_primary"]')?.checked;

            const data = {
                id: id ? parseInt(id) : null,
                contact_id: parseInt(contactId),
                email: currentEmail,
                label_type_id: currentType,
                is_primary: currentPrimary
            };

            if (id) {
                // It's an existing record - Update it if 'dirty'
                const isDirty = 
                    currentEmail !== row.getAttribute('data-original-email') ||
                    currentType !== parseInt(row.getAttribute('data-original-type')) || 
                    currentPrimary !== (row.getAttribute('data-original-primary') === 'true');

                if (isDirty) {
                    requests.push(fetch(`/api/v1/emails/${id}`, {
                        method: 'PATCH',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(data)
                    }));
                }
            } else {
                // It's a new record - Create it
                requests.push(fetch(`/api/v1/contacts/${contactId}/emails`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                }));
            }
        });

        // 3. Execute all requests concurrently
        try {
            const results = await Promise.all(requests);
            const allOk = results.every(res => res.ok);

            if (allOk) {
                showNotification('Email records synced!', 'success');
                deletedEmailIds = []; // Clear the delete queue
                setTimeout(() => location.reload(), 500);
            } else {
                showNotification('Some updates failed.', 'error');
                btn.disabled = false;
            }
        } catch (err) {
            console.error("Sync error:", err);
            showNotification('Network error during sync', 'error');
            btn.disabled = false;
        }
    };

    /////////////////////
    // end email handling
    /////////////////////

    /////////////////////
    // Add phone handling
    /////////////////////

    window.markPhonesChanged = function() {
        document.getElementById('savePhonesBtn').disabled = false;
        document.getElementById('savePhonesBtn').classList.add('btn-pulse'); // Optional flair
    };

    window.addPhone = function() {
        const container = document.getElementById('phonesList');
        // Get the options from our hidden template
        const optionsHtml = document.getElementById('phoneTypeOptionsTmpl').innerHTML;
        
        const div = document.createElement('div');
        div.className = 'flex gap-2 items-start phone-row';  // Note: No data-id attribute means it's a new record
        
        div.innerHTML = `
            <input type="tel" name="phone" placeholder="(555) 123-4567" class="input input-bordered flex-1" required>
            <select name="label_type_id" class="select select-bordered">
                ${optionsHtml}
            </select>
            <label class="label cursor-pointer gap-2">
                <input type="checkbox" name="is_primary" class="checkbox checkbox-primary">
                <span class="label-text">Primary</span>
            </label>
            <button type="button" class="btn btn-ghost btn-sm btn-square" onclick="removePhoneRow(this)">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        `;
        
        container.appendChild(div);
        markPhonesChanged();
    };

    let deletedPhoneIds = [];

    window.removePhoneRow = function(btn) {
        const row = btn.closest('.phone-row');
        const id = row.getAttribute('data-id');
        
        if (id) {
            deletedPhoneIds.push(parseInt(id));
        }
        
        row.remove();
        markPhonesChanged();
    };

    window.savePhonesOnly = async function() {
        const btn = document.getElementById('savePhonesBtn');
        const contactId = window.location.pathname.split('/').pop();
        const phoneRows = document.querySelectorAll('.phone-row');
        
        btn.disabled = true;
        const requests = [];

        // 1. Handle Deletions first
        deletedPhoneIds.forEach(id => {
            requests.push(fetch(`/api/v1/contacts/${contactId}/phones/${id}`, { method: 'DELETE' }));
        });

        // 2. Loop through rows to determine POST (new) vs PATCH (update)
        phoneRows.forEach(row => {
            const id = row.getAttribute('data-id');

            const currentPhone = row.querySelector('[name="phone"]').value;
            const currentType = parseInt(row.querySelector('[name="label_type_id"]').value);
            const currentPrimary = !!row.querySelector('[name="is_primary"]')?.checked;

            const data = {
                id: id ? parseInt(id) : null,
                contact_id: parseInt(contactId),
                phone: currentPhone,
                label_type_id: currentType,
                is_primary: currentPrimary
            };

            if (id) {
                // It's an existing record - Update it if 'dirty'
                const isDirty = 
                    currentPhone !== row.getAttribute('data-original-phone') ||
                    currentType !== parseInt(row.getAttribute('data-original-type')) || 
                    currentPrimary !== (row.getAttribute('data-original-primary') === 'true');

                if (isDirty) {
                    requests.push(fetch(`/api/v1/phones/${id}`, {
                        method: 'PATCH',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(data)
                    }));
                }
            } else {
                // It's a new record - Create it
                requests.push(fetch(`/api/v1/contacts/${contactId}/phones`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                }));
            }
        });

        // 3. Execute all requests concurrently
        try {
            const results = await Promise.all(requests);
            const allOk = results.every(res => res.ok);

            if (allOk) {
                showNotification('Phone records synced!', 'success');
                deletedPhoneIds = []; // Clear the delete queue
                setTimeout(() => location.reload(), 500);
            } else {
                showNotification('Some updates failed.', 'error');
                btn.disabled = false;
            }
        } catch (err) {
            console.error("Sync error:", err);
            showNotification('Network error during sync', 'error');
            btn.disabled = false;
        }
    };

    /////////////////////
    // end phone handling
    /////////////////////

    /////////////////////
    // Add address handling
    /////////////////////

    window.markAddressesChanged = function() {
        document.getElementById('saveAddressesBtn').disabled = false;
        document.getElementById('saveAddressesBtn').classList.add('btn-pulse'); // Optional flair
    };

    // Add address field
    window.addAddress = function() {
        const container = document.getElementById('addressesList');
        const optionsHtml = document.getElementById('addressTypeOptionsTmpl').innerHTML;
        
        const div = document.createElement('div');
        div.className = 'card bg-base-200 shadow-md p-4 address-row'; // Note: No data-id attribute means it's a new record

        div.innerHTML = `
            <div class="flex justify-between items-start mb-3">
                <select name="label_type_id" class="select select-bordered select-sm">
                    ${optionsHtml}
                </select>
                <div class="flex gap-2">
                    <label class="label cursor-pointer gap-2">
                        <input type="checkbox" name="is_primary" class="checkbox checkbox-primary checkbox-sm">
                        <span class="label-text">Primary</span>
                    </label>
                    <button type="button" class="btn btn-ghost btn-sm btn-square" onclick="removeAddressRow(this)">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>
            </div>
            <div class="space-y-2">
                <div class="grid grid-cols-5 gap-2">
                    <input type="text" name="street" placeholder="Street Address" class="input input-bordered input-sm col-span-3">
                    <input type="text" name="extended_street" placeholder="Apt/Ste" class="input input-bordered input-sm col-span-2">
                </div>

                <div class="grid grid-cols-5 gap-2">
                    <input type="text" name="city" placeholder="City" class="input input-bordered input-sm col-span-3">
                    <input type="text" name="state" placeholder="State" class="input input-bordered input-sm col-span-2">
                </div>

                <div class="grid grid-cols-2 gap-2">
                    <input type="text" name="postal_code" placeholder="ZIP" class="input input-bordered input-sm">
                    <input type="text" name="country" placeholder="Country" class="input input-bordered input-sm">
                </div>
            </div>
        `;
        
        container.appendChild(div);
        markAddressesChanged();
    };

    let deletedAddressIds = [];

    window.removeAddressRow = function(btn) {
        const row = btn.closest('.address-row');
        const id = row.getAttribute('data-id');
        
        if (id) {
            deletedAddressIds.push(parseInt(id));
        }
        
        row.remove();
        markAddressesChanged();
    };

    window.saveAddressesOnly = async function() {
        const btn = document.getElementById('saveAddressesBtn');
        const contactId = window.location.pathname.split('/').pop();
        const addressRows = document.querySelectorAll('.address-row');
        
        btn.disabled = true;
        const requests = [];

        // 1. Handle Deletions first
        deletedAddressIds.forEach(id => {
            requests.push(fetch(`/api/v1/contacts/${contactId}/addresses/${id}`, { method: 'DELETE' }));
        });

        // 2. Loop through rows to determine POST (new) vs PATCH (update)
        addressRows.forEach(row => {
            const id = row.getAttribute('data-id');

            const currentStreet = row.querySelector('[name="street"]').value;
            const currentExtStreet = row.querySelector('[name="extended_street"]').value;
            const currentCity = row.querySelector('[name="city"]').value;
            const currentState = row.querySelector('[name="state"]').value;
            const currentPostal = row.querySelector('[name="postal_code"]').value;
            const currentCountry = row.querySelector('[name="country"]').value;
            const currentType = parseInt(row.querySelector('[name="label_type_id"]').value);
            const currentPrimary = !!row.querySelector('[name="is_primary"]')?.checked;

            const data = {
                id: id ? parseInt(id) : null,
                contact_id: parseInt(contactId),
                street: currentStreet,
                extended_street: currentExtStreet,
                city: currentCity,
                state: currentState,
                postal_code: currentPostal,
                country: currentCountry,
                label_type_id: currentType,
                is_primary: currentPrimary
            };

            if (id) {
                // It's an existing record - Update it if 'dirty
                const isDirty =
                    currentStreet !== row.getAttribute('data-original-street') ||
                    currentExtStreet !== row.getAttribute('data-original-extstreet') ||
                    currentCity !== row.getAttribute('data-original-city') ||
                    currentState !== row.getAttribute('data-original-state') ||
                    currentPostal !== row.getAttribute('data-original-postal') ||
                    currentCountry  !== row.getAttribute('data-original-country') ||
                    currentType !== parseInt(row.getAttribute('data-original-type')) || 
                    currentPrimary !== (row.getAttribute('data-original-primary') === 'true');

                if (isDirty) {
                    requests.push(fetch(`/api/v1/addresses/${id}`, {
                        method: 'PATCH',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(data)
                    }));
                }
            } else {
                // It's a new record - Create it
                requests.push(fetch(`/api/v1/contacts/${contactId}/addresses`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                }));
            }
        });

        // 3. Execute all requests concurrently
        try {
            const results = await Promise.all(requests);
            const allOk = results.every(res => res.ok);

            if (allOk) {
                showNotification('Address records synced!', 'success');
                deletedAddressIds = []; // Clear the delete queue
                setTimeout(() => location.reload(), 500);
            } else {
                showNotification('Some updates failed.', 'error');
                btn.disabled = false;
            }
        } catch (err) {
            console.error("Sync error:", err);
            showNotification('Network error during sync', 'error');
            btn.disabled = false;
        }
    };

    /////////////////////
    // end address handling
    /////////////////////

    /////////////////////
    // Add organization handling
    /////////////////////

    window.markOrganizationsChanged = function() {
        document.getElementById('saveOrganizationsBtn').disabled = false;
        document.getElementById('saveOrganizationsBtn').classList.add('btn-pulse'); // Optional flair
    };

    // Add organization field
    window.addOrganization = function() {
        const container = document.getElementById('orgsList');
                
        const div = document.createElement('div');
        div.className = 'card bg-base-200 shadow-md p-4 organization-row';

        div.innerHTML = `
            <div class="flex gap-2 items-center mb-3">
                <input type="text" name="company_name" placeholder="Company Name" class="input input-bordered input-sm w-[47%]">
                    
                <input type="text" name="job_title" placeholder="Job Title" class="input input-bordered input-sm w-[47%]">
                    
                <button type="button" class="btn btn-ghost btn-sm btn-square ml-auto" onclick="removeOrganizationRow(this)">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
            </div>

            <div class="w-full">
                <input type="text" name="department" placeholder="Department" class="input input-bordered input-sm w-full">
            </div>
        `;
        
        container.appendChild(div);
        markOrganizationsChanged();
    };

    let deletedOrganizationIds = [];

    window.removeOrganizationRow = function(btn) {
        const row = btn.closest('.organization-row');
        const id = row.getAttribute('data-id');
        
        if (id) {
            deletedOrganizationIds.push(parseInt(id));
        }
        
        row.remove();
        markOrganizationsChanged();
    };

    window.saveOrganizationsOnly = async function() {
        const btn = document.getElementById('saveOrganizationsBtn');
        const contactId = window.location.pathname.split('/').pop();
        const organizationRows = document.querySelectorAll('.organization-row');
        
        btn.disabled = true;
        const requests = [];

        // 1. Handle Deletions first
        deletedAddressIds.forEach(id => {
            requests.push(fetch(`/api/v1/contacts/${contactId}/organizations/${id}`, { method: 'DELETE' }));
        });

        // 2. Loop through rows to determine POST (new) vs PATCH (update)
        organizationRows.forEach(row => {
            const id = row.getAttribute('data-id');

            const currentCompanyName = row.querySelector('[name="company_name"]').value;
            const currentJobTitle = row.querySelector('[name="job_title"]').value;
            const currentDepartment = row.querySelector('[name="department"]').value;

            const data = {
                id: id ? parseInt(id) : null,
                contact_id: parseInt(contactId),
                name: currentCompanyName,
                title: currentJobTitle,
                department: currentDepartment,
            };

            if (id) {
                // It's an existing record - Update it if 'dirty
                const isDirty =
                    currentCompanyName !== row.getAttribute('data-original-companyname') ||
                    currentJobTitle !== row.getAttribute('data-original-jobtitle') ||
                    currentDepartment !== row.getAttribute('data-original-department')

                if (isDirty) {
                    requests.push(fetch(`/api/v1/organizations/${id}`, {
                        method: 'PATCH',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(data)
                    }));
                }
            } else {
                // It's a new record - Create it
                requests.push(fetch(`/api/v1/contacts/${contactId}/organizations`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                }));
            }
        });

        // 3. Execute all requests concurrently
        try {
            const results = await Promise.all(requests);
            const allOk = results.every(res => res.ok);

            if (allOk) {
                showNotification('Organization records synced!', 'success');
                deletedAddressIds = []; // Clear the delete queue
                setTimeout(() => location.reload(), 500);
            } else {
                showNotification('Some updates failed.', 'error');
                btn.disabled = false;
            }
        } catch (err) {
            console.error("Sync error:", err);
            showNotification('Network error during sync', 'error');
            btn.disabled = false;
        }
    };

    /////////////////////
    // end organization handling
    /////////////////////

    /////////////////////
    // Add url handling
    /////////////////////

    window.markURLsChanged = function() {
        document.getElementById('saveURLsBtn').disabled = false;
        document.getElementById('saveURLsBtn').classList.add('btn-pulse'); // Optional flair
    };

    // Add URL field
    window.addURL = function() {
        const container = document.getElementById('urlsList');
        const optionsHtml = document.getElementById('urlTypeOptionsTmpl').innerHTML;
        
        const div = document.createElement('div');
        div.className = 'flex gap-2 items-center url-row';

        div.innerHTML = `
            <input type="url" name="url" placeholder="https://example.com" class="input input-bordered input-sm flex-1">
            <select name="label_type_id" class="select select-bordered">
                ${optionsHtml}
            </select>
            <button type="button" class="btn btn-ghost btn-sm btn-square" onclick="removeURLRow(this)">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        `;
        
        container.appendChild(div);
        markURLsChanged();
    };

    let deletedURLIds = [];

    window.removeURLRow = function(btn) {
        const row = btn.closest('.url-row');
        const id = row.getAttribute('data-id');
        
        if (id) {
            deletedURLIds.push(parseInt(id));
        }
        
        row.remove();
        markURLsChanged();
    };

    window.saveURLsOnly = async function() {
        const btn = document.getElementById('saveURLsBtn');
        const contactId = window.location.pathname.split('/').pop();
        const urlRows = document.querySelectorAll('.url-row');
        
        btn.disabled = true;
        const requests = [];

        // 1. Handle Deletions first
        deletedURLIds.forEach(id => {
            requests.push(fetch(`/api/v1/contacts/${contactId}/urls/${id}`, { method: 'DELETE' }));
        });

        // 2. Loop through rows to determine POST (new) vs PATCH (update)
        urlRows.forEach(row => {
            const id = row.getAttribute('data-id');

            const currentURL = row.querySelector('[name="url"]').value;
            const currentType = parseInt(row.querySelector('[name="label_type_id"]').value);
            const currentPrimary = !!row.querySelector('[name="is_primary"]')?.checked;

            const data = {
                id: id ? parseInt(id) : null,
                contact_id: parseInt(contactId),
                url: currentURL,
                label_type_id: currentType,
                is_primary: currentPrimary
            };

            if (id) {
                // It's an existing record - Update it if 'dirty'
                const isDirty = 
                    currentURL !== row.getAttribute('data-original-url') ||
                    currentType !== parseInt(row.getAttribute('data-original-type')) || 
                    currentPrimary !== (row.getAttribute('data-original-primary') === 'true');

                if (isDirty) {
                    requests.push(fetch(`/api/v1/urls/${id}`, {
                        method: 'PATCH',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(data)
                    }));
                }
            } else {
                // It's a new record - Create it
                requests.push(fetch(`/api/v1/contacts/${contactId}/urls`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                }));
            }
        });

        // 3. Execute all requests concurrently
        try {
            const results = await Promise.all(requests);
            const allOk = results.every(res => res.ok);

            if (allOk) {
                showNotification('Website records synced!', 'success');
                deletedURLIds = []; // Clear the delete queue
                setTimeout(() => location.reload(), 500);
            } else {
                showNotification('Some updates failed.', 'error');
                btn.disabled = false;
            }
        } catch (err) {
            console.error("Sync error:", err);
            showNotification('Network error during sync', 'error');
            btn.disabled = false;
        }
    };

    /////////////////////
    // end url handling
    /////////////////////

    /////////////////////
    // add notes handling
    /////////////////////

    window.markNotesChanged = function() {
        document.getElementById('saveNotesBtn').disabled = false;
        document.getElementById('saveNotesBtn').classList.add('btn-pulse'); // Optional flair
    };

    window.saveNotesOnly = async function() {
        const btn = document.getElementById('saveNotesBtn');
        const notesTextArea = document.getElementById('notesField');
        const contactId = window.location.pathname.split('/').pop();

        // Check if it's actually changed before doing anything
        const isDirty = notesTextArea.value !== notesTextArea.getAttribute('data-original');
        if (!isDirty) return;

        // UI Feedback: Loading state
        btn.disabled = true;

        const data = {
            contact_id: parseInt(contactId),
            notes: notesTextArea.value
        };

        try {
            const response = await fetch(`/api/v1/contacts/${contactId}/notes`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });

            if (response.ok) {
                showNotification('Notes saved!', 'success');              
                setTimeout(() => location.reload(), 500);
            } else {
                throw new Error('Failed to save notes');
            }
        } catch (err) {
            console.error("Save error:", err);
            showNotification('Failed to save notes', 'error');
            btn.disabled = false;
        }
    };

    /////////////////////
    // end notes handling
    /////////////////////

    /////////////////////
    // add dates handling
    /////////////////////

    window.markDatesChanged = function() {
        document.getElementById('saveDatesBtn').disabled = false;
        document.getElementById('saveDatesBtn').classList.add('btn-pulse');  // Optional flair
    };

    // Add Other Date field
    window.addOtherDate = function() {
        const container = document.getElementById('otherDatesList');
        
        const div = document.createElement('div');
        div.className = 'flex gap-2 items-start border-l-4 border-primary pl-3 other-date-row';

        div.innerHTML = `
            <div class="flex-1 space-y-2">
                <input type="text" name="event_name" placeholder="Event name" class="input input-bordered input-sm w-full" required>
                <div class="grid gap-2" style="grid-template-columns: 1fr 0.8fr 0.8fr;">
                    <select name="event_date_month" class="select select-bordered select-sm">
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
                    <select name="event_date_day" class="select select-bordered select-sm">
                        <option value="">*Day</option>
                        ${Array.from({length: 31}, (_, i) => `<option value="${i+1}">${i+1}</option>`).join('')}
                    </select>
                    <input type="number" name="event_date_year" placeholder="Year" min="1900" max="2100" class="input input-bordered input-sm">
                </div>
            </div>
            <button type="button" class="btn btn-ghost btn-xs btn-circle" onclick="removeOtherDateRow(this)" title="Delete event">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        `;
        
        container.appendChild(div);        
        markDatesChanged();
    };

    let deletedOtherDateIds = [];

    window.removeOtherDateRow = function(btn) {
        const row = btn.closest('.other-date-row');
        const id = row.getAttribute('data-id');
        
        if (id) {
            deletedOtherDateIds.push(parseInt(id));
        }
        
        row.remove();
        markDatesChanged();
    };

    window.saveDatesOnly = async function() {
        const btn = document.getElementById('saveDatesBtn');
        const contactId = window.location.pathname.split('/').pop();
        const requests = [];

        btn.disabled = true;
        btn.innerHTML = '<span class="loading loading-spinner loading-xs"></span> Saving...';

        // 1. Birthday PATCH (if dirty)
        if (isGroupDirty('birthday', document.getElementById('datesCard'))) {
            requests.push(fetch(`/api/v1/contacts/${contactId}/birthday`, {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(formatEndpointBody(getGroupValues(document.getElementById('datesCard'), 'birthday')))
            }));
        }

        // 2. Anniversary PATCH (if dirty)
        if (isGroupDirty('anniversary', document.getElementById('datesCard'))) {
            requests.push(fetch(`/api/v1/contacts/${contactId}/anniversary`, {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(formatEndpointBody(getGroupValues(document.getElementById('datesCard'), 'anniversary')))
            }));
        }

        // 3. Other Dates: Updates & Creations
        document.querySelectorAll('.other-date-row').forEach(row => {
            const id = row.getAttribute('data-id');
            const data = {
                date_type: row.querySelector('[name="event_name"]').value,
                ...formatEndpointBody(getGroupValues(row, 'event_date'))
            };

            if (id) {
                if (isOtherDateRowDirty(row)) {
                    requests.push(fetch(`/api/v1/other-dates/${id}`, {
                        method: 'PATCH',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify(data)
                    }));
                }
            } else {
                // Creation
                requests.push(fetch(`/api/v1/contacts/${contactId}/other-dates`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                }));
            }
        });

        // 4. Deletions
        deletedOtherDateIds.forEach(id => {
            requests.push(fetch(`/api/v1/contacts/${contactId}/other-dates/${id}`, { method: 'DELETE' }));
        });

        try {
            const results = await Promise.all(requests);
            if (results.every(r => r.ok)) {
                showNotification('All dates updated!', 'success');
                deletedOtherDateIds = []; // Clear the delete queue
                setTimeout(() => location.reload(), 500);
            } else {
                showNotification('Some updates failed.', 'error');
                btn.disabled = false;
            }
        } catch (err) {
            showNotification('Network error', 'error');
            btn.disabled = false;
        }
    };

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
        markDatesChanged();
    };

    function getGroupValues(container, prefix = "") {
        const p = prefix ? prefix + "_" : "";
        return {
            name:  container.querySelector(`[name="${p}name"]`)?.value || 
                container.querySelector(`[name="${p}event_name"]`)?.value,
            month: container.querySelector(`[name="${p}month"]`)?.value,
            day:   container.querySelector(`[name="${p}day"]`)?.value,
            year:  container.querySelector(`[name="${p}year"]`)?.value
        };
    }

    function formatEndpointBody(vals) {
        // if the date birthday or anniversary has been cleared out
        if (!vals.month || !vals.day) {
            return {
                date: null,
                date_month: null,
                date_day: null
            };
        }

        // Logic: If year exists, send ISO string. If not, send components.
        if (vals.year && vals.year > 0) {
            // Construct YYYY-MM-DD. Using UTC to avoid timezone shifts.
            const dateStr = `${vals.year}-${String(vals.month).padStart(2, '0')}-${String(vals.day).padStart(2, '0')}`;
            return {
                date: new Date(dateStr).toISOString()
            };
        } else {
            return {
                date_month: parseInt(vals.month),
                date_day: parseInt(vals.day)
            };
        }
    }

    function isGroupDirty(type, container) {
        const clean = (val) => (!val || val === "0" || val === 0) ? "" : String(val).trim();

        const m = clean(document.getElementById(`${type}_month`).value);
        const d = clean(document.getElementById(`${type}_day`).value);
        const y = clean(document.getElementById(`${type}_year`).value);

        const om = clean(container.getAttribute(`data-original-${type}-month`));
        const od = clean(container.getAttribute(`data-original-${type}-day`));
        const oy = clean(container.getAttribute(`data-original-${type}-year`));

        return m !== om || d !== od || y !== oy;
    }

    // Helper: Check Other Date Row baseline
    function isOtherDateRowDirty(row) {
        const clean = (val) => (!val || val === "0" || val === 0) ? "" : String(val).trim();

        const currentName  = clean(row.querySelector('[name="event_name"]').value);
        const currentMonth = clean(row.querySelector('[name="event_date_month"]').value);
        const currentDay   = clean(row.querySelector('[name="event_date_day"]').value);
        const currentYear  = clean(row.querySelector('[name="event_date_year"]').value);

        const originalName  = clean(row.getAttribute('data-original-name'));
        const originalMonth = clean(row.getAttribute('data-original-month'));
        const originalDay   = clean(row.getAttribute('data-original-day'));
        const originalYear  = clean(row.getAttribute('data-original-year'));

        return currentName  !== originalName ||
            currentMonth !== originalMonth ||
            currentDay   !== originalDay ||
            currentYear  !== originalYear;
    }

    /////////////////////
    // end dates handling
    /////////////////////

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

    window.toggleMaidenField = function() {
        const gender = document.getElementById('genderSelect').value;
        const container = document.getElementById('maidenFieldContainer');

        if (gender === 'M') {
            container.classList.add('hidden');
        } else {
            container.classList.remove('hidden');
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

    console.log('Contact detail page initialized');
})();
