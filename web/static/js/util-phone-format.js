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

    // Hook for the Common Script to update preview when card rotates
    window.onCardChange = function() {
        window.previewPhone();
    };

    window.previewPhone = function() {
        const selector = document.getElementById('format-selector');
        const card = document.querySelector('.util-card:not(.hidden)');
        
        if (!selector || !card) return;

        const pattern = selector.value;
        const raw = card.getAttribute('data-raw');
        const previewEl = card.querySelector('.preview-text');

        if (previewEl && raw) {
            previewEl.innerText = formatString(raw, pattern);
        }
    };

    function formatString(raw, pattern) {
        // Strip everything but digits
        const digits = raw.replace(/\D/g, "");
        let clean = digits;

        // Handle US country code if present
        if (digits.length === 11 && digits.startsWith("1")) {
            clean = digits.substring(1);
        }

        // Only format if we have exactly 10 digits
        if (clean.length !== 10) return raw;

        let formatted = pattern;
        for (let i = 0; i < 10; i++) {
            formatted = formatted.replace('X', clean[i]);
        }
        return formatted;
    }

    window.applyFormat = async function(event, contactId, phoneId) {
        const btn = event.currentTarget;
        const card = btn.closest('.util-card');
        const newPhone = card.querySelector('.preview-text').innerText;

        btn.disabled = true;
        btn.innerHTML = '<span class="loading loading-spinner loading-xs"></span>';

        try {
            // This must match your API route for updating a specific phone record
            const response = await fetch(`/api/v1/phones/${phoneId}`, {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ phone: newPhone })
            });

            if (response.ok) {
                UtilCommon.showNext(card); // This handles the removal and showing the next card
            } else {
                btn.disabled = false;
                btn.innerText = "Apply Format";
                console.error("Save failed");
            }
        } catch (e) {
            btn.disabled = false;
            btn.innerText = "Apply Format";
            console.error(e);
        }
    };

})();