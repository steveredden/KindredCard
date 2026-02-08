/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

// KindredCard - Utilities page: Phone Standardization
(function() {
    'use strict';
    
    window.linkImmich = async function(event) {
        const btn = event.currentTarget;
        const card = btn.closest('.util-card');
        const contactId = card.dataset.contactId;
        const personId = card.dataset.personId;
        
        btn.disabled = true;
        btn.innerHTML = '<span class="loading loading-spinner loading-xs"></span> Linking...';

        try {
            // 1. Create the link in KindredCard
            const linkRes = await fetch('/api/v1/immich/link', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    contact_id: parseInt(contactId),
                    person_id: personId
                })
            });

            if (!linkRes.ok) throw new Error("Failed to create link");

            // 2. Trigger a sync (optional: you can create a specific endpoint for one person)
            // For now, we move to the next card
            UtilCommon.showNext(card);

        } catch (err) {
            console.error(err);
            btn.disabled = false;
            btn.innerText = "Error - Try Again";
        }
    }

    window.unlinkImmich = async function(contactId, urlId) {
        if (!confirm("Are you sure you want to remove the Immich link for this contact?")) return;

        try {
            // We delete the URL record that has the 'immich' type
            const res = await fetch(`/api/v1/contacts/${contactId}/urls/${urlId}`, {
                method: 'DELETE'
            });
            if (res.ok) window.location.reload();
        } catch (err) {
            alert("Failed to unlink");
        }
    }

    window.syncField = async function(contactId, field, btn) {
        btn.classList.add('loading');
        
        let url = `/api/v1/contacts/${contactId}`;
        let body = {};

        try {
            if (field === "birthday") {
                // Pull the date string directly from the Immich column in the UI
                const row = btn.closest('tr');
                const immichDate = row.querySelector('.immich-birthdate').innerText.trim();

                if (!immichDate || immichDate === "No Birthday Saved") {
                    btn.classList.remove('loading');
                    console.log("No birthday to sync! Skipping!")
                    return;
                }
                
                url = `/api/v1/contacts/${contactId}/birthday`;
                body = {
                    date: new Date(immichDate).toISOString()
                };
            } 
            else if (field === "avatar") {
                // 1. Get the Proxy URL from the Immich Avatar in this row
                const row = btn.closest('tr');
                const imgUrl = row.querySelector('.immich-avatar-img').src;

                // 2. Fetch the image bytes and convert to Base64
                const response = await fetch(imgUrl);
                const blob = await response.blob();
                const base64 = await new Promise((resolve) => {
                    const reader = new FileReader();
                    reader.onloadend = () => resolve(reader.result.split(',')[1]);
                    reader.readAsDataURL(blob);
                });

                body = {
                    avatar_base64: base64,
                    avatar_mime_type: blob.type
                };
            }

            const res = await fetch(url, {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body)
            });

            if (res.ok) {
                btn.classList.remove('loading');
                btn.innerHTML = '✅';
                setTimeout(() => window.location.reload(), 200);
            }
        } catch (err) {
            console.error(err);
            btn.classList.remove('loading');
            alert("Sync failed");
        }
    };

})();