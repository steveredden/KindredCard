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

    // 1. Define the function first
    const previewAll = function() {
        const selector = document.getElementById('format-selector');
        if (!selector) return;

        const pattern = selector.value;
        const cards = document.querySelectorAll('.phone-card');
        
        cards.forEach(card => {
            const raw = card.getAttribute('data-raw');
            const previewEl = card.querySelector('.preview-text');
            if (previewEl && raw) {
                previewEl.innerText = window.formatNumber(raw, pattern);
            }
        });
    };

    // 2. Attach to window so HTML 'onchange' can see it
    window.previewAll = previewAll;

    window.formatNumber = function(raw, pattern) {
        if (!raw) return "";
        const digits = raw.replace(/\D/g, "");
        let clean = digits;
        if (digits.length === 11 && digits.startsWith("1")) {
            clean = digits.substring(1);
        }
        if (clean.length !== 10) return raw;

        let formatted = pattern;
        for (let i = 0; i < 10; i++) {
            formatted = formatted.replace("X", clean[i]);
        }
        return formatted;
    };

    // 3. Now add the listener using the locally defined variable
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', previewAll);
    } else {
        previewAll();
    }

    window.applyFormat = async function() {
        const currentCard = document.querySelector('.phone-card:not(.hidden)');
        const phoneId = currentCard.getAttribute('data-phone-id');
        const newFormat = currentCard.querySelector('.preview-text').innerText;

        // UI: Instant transition
        showNextCard(currentCard);

        try {
            const response = await fetch(`/api/v1/phones/${phoneId}`, {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ phone: newFormat })
            });

            if (!response.ok) {
                console.error("Failed to update phone ID:", phoneId);
            }
        } catch (err) {
            console.error("Network error updating phone:", err);
        }
    }

    window.skipPhone = function() {
        const currentCard = document.querySelector('.phone-card:not(.hidden)');
        showNextCard(currentCard);
    }

    window.showNextCard = function(currentCard) {
        const nextCard = currentCard.nextElementSibling;
        currentCard.remove();

        // Decrement the UI counter
        window.decrementBadge();

        if (nextCard) {
            nextCard.classList.remove('hidden');
            window.previewAll(); 
        } else {
            document.getElementById('contact-deck').innerHTML = `
                <div class="text-center py-20 animate-bounce">
                    <h3 class="text-xl font-bold text-success">Done!</h3>
                    <p>You've cleared the queue!</p>
                    <p>Refresh for more</p>
                </div>`;
        }
    }

    window.decrementBadge = function() {
        const badge = document.getElementById('remaining-count');
        if (badge) {
            // Extract the number from "100 remaining" or similar text
            const currentText = badge.innerText;
            const match = currentText.match(/\d+/);
            if (match) {
                const count = parseInt(match[0], 10);
                const newCount = Math.max(0, count - 1);
                badge.innerText = `${newCount} remaining`;
            }
        }
    };
})();