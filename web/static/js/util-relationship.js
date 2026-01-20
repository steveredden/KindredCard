/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

// KindredCard - Utilities page: Relationship Suggestions
(function() {
    'use strict';
    
    window.applyRelationship = async function(targetID, sourceID, typeId) {
        const btn = event.currentTarget;
        const card = btn.closest('.suggestion-card');
        
        btn.disabled = true;
        btn.innerHTML = '<span class="loading loading-spinner loading-xs"></span> linking...';

        try {
            // vCard Flow: POST to the Source (Leah) 
            // to say she has a "Daughter" (Target: Sarah)
            const response = await fetch(`/api/v1/contacts/${sourceID}/relationships`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    related_contact_id: parseInt(targetID),
                    relationship_type_id: parseInt(typeId)
                })
            });

            if (response.ok) {
                // UI advancement logic
                card.classList.add('hidden');
                const nextCard = card.nextElementSibling;
                if (nextCard && nextCard.classList.contains('suggestion-card')) {
                    nextCard.classList.remove('hidden');
                } else {
                    window.location.reload(); 
                }
                updateCount();
            } else {
                const txt = await response.text();
                alert("Error: " + txt);
                btn.disabled = false;
                btn.innerText = "Link Contacts";
            }
        } catch (e) {
            console.error(e);
            btn.disabled = false;
        }
    };

    window.skipSuggestion = function() {
        const currentCard = document.querySelector('.suggestion-card:not(.hidden), .anniversary-card:not(.hidden)');
        if (currentCard) {
            currentCard.classList.add('hidden');
            const nextCard = currentCard.nextElementSibling;
            if (nextCard && (nextCard.classList.contains('suggestion-card') || nextCard.classList.contains('anniversary-card'))) {
                nextCard.classList.remove('hidden');
            } else {
                // Reached end of batch
                window.location.reload(); 
            }
            updateCount();
        }
    }

    window.updateCount = function() {
        const countBadge = document.getElementById('remaining-count');
        const remaining = document.querySelectorAll('.suggestion-card:not(.hidden), .anniversary-card:not(.hidden)').length;
        countBadge.innerText = `${remaining} remaining`;
    }

})();