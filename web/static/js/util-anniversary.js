/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

// KindredCard - Utilities page: Annivesary Patching
(function() {
    'use strict';
    
    window.syncAnniversary = async function(event, targetId, sourceId, displayVal) {
        // Use event.currentTarget to get the button regardless of where the click hit
        const btn = event.currentTarget;
        const card = document.getElementById(`suggestion-event-${targetId}-${sourceId}`);
        
        btn.disabled = true;
        const originalText = btn.innerHTML;
        btn.innerHTML = '<span class="loading loading-spinner loading-xs"></span> Syncing...';

        try {
            let payload = {};
            
            // Clean the string in case any extra text snuck in
            // Expecting "YYYY-MM-DD" or "MM-DD"
            const parts = displayVal.split('-');
            
            if (parts.length === 3) {
                // Full Date: YYYY-MM-DD
                payload.anniversary = displayVal + "T00:00:00Z";
            } else if (parts.length === 2) {
                // Partial Date: MM-DD
                payload.anniversary_month = parseInt(parts[0]);
                payload.anniversary_day = parseInt(parts[1]);
            } else {
                throw new Error("Received unexpected date format: " + displayVal);
            }

            console.log("Attempting Sync with payload:", payload);

            const response = await fetch(`/api/v1/contacts/${targetId}/anniversary`, {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });

            if (response.ok) {
                // Hide current card
                card.classList.add('hidden');
                
                // Find and show next card
                const nextCard = card.nextElementSibling;
                if (nextCard && nextCard.classList.contains('anniversary-card')) {
                    nextCard.classList.remove('hidden');
                } else {
                    // No more cards left, reload to show the "Success" state
                    window.location.reload(); 
                }
                updateCount();
            } else {
                const errorText = await response.text();
                throw new Error(errorText || "Server returned " + response.status);
            }
        } catch (error) {
            console.error("Sync Error:", error);
            alert('Error: ' + error.message);
            btn.disabled = false;
            btn.innerHTML = originalText;
        }
    };

    window.skipSuggestion = function() {
        const currentCard = document.querySelector('.anniversary-card:not(.hidden)');
        if (currentCard) {
            currentCard.classList.add('hidden');
            const nextCard = currentCard.nextElementSibling;
            if (nextCard && nextCard.classList.contains('anniversary-card')) {
                nextCard.classList.remove('hidden');
            } else {
                window.location.reload(); 
            }
            updateCount();
        }
    };

    window.updateCount = function() {
        const countBadge = document.getElementById('remaining-count');
        const remaining = document.querySelectorAll('.anniversary-card:not(.hidden)').length;
        if (countBadge) {
            countBadge.innerText = `${remaining} remaining`;
        }
    };
})();