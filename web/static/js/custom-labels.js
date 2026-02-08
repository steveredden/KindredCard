/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

// KindredCard - Custom Labels page
(function() {
    'use strict';

    window.saveLabel = async function(event) {
        event.preventDefault();
        const formData = new FormData(event.target);
        const payload = Object.fromEntries(formData.entries());

        try {
            const res = await fetch('/api/v1/settings/labels', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });

            if (res.ok) {
                add_label_modal.close()
                showNotification('Label created', 'success');
                setTimeout(() => location.reload(), 500);
            } else {
                const err = await res.json();
                alert(err.message || 'Error creating label');
                showNotification('Label creation failed', 'error');
            }
        } catch (e) {
            showNotification('Label creation failed', 'error');
            console.error(e);
        }
    }

    window.deleteLabel = async function(id) {
        if (!confirm('Are you sure you want to delete this custom label?')) return;

        try {
            const res = await fetch(`/api/v1/settings/labels/${id}`, {
                method: 'DELETE'
            });
            if (res.ok) {
                showNotification('Label deleted', 'success');
                setTimeout(() => location.reload(), 500);
            }
        } catch (e) {
            showNotification('Label deletion failed', 'error');
            console.error(e);
        }
    }

    window.openAddLabelModal = function() {
        document.getElementById('newLabelForm').reset();
        add_label_modal.showModal();
    }

    console.log('Custom Label page initialized');
})();