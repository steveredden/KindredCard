/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

// KindredCard - Utilities page: Gender Assignment
(function() {
    'use strict';
    
    window.assignGender = function(gender) {
        const currentCard = document.querySelector('.contact-card:not(.hidden)');
        if (!currentCard) return;

        const contactId = currentCard.getAttribute('data-id');

        // Fast-path: Hide current and show next immediately for UX
        showNextCard(currentCard);

        if (gender !== 'skip') {
            fetch(`/api/v1/contacts/${contactId}`, {
                method: 'PATCH',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ gender: gender })
            }).catch(err => console.error('Failed to update gender:', err));
        }
    }

    window.deleteContact = async function() {
        const currentCard = document.querySelector('.contact-card:not(.hidden)');
        if (!currentCard) return;
        
        const confirmed = await customConfirm("Are you sure you want to delete this contact?")
        if (!confirmed) return;

        const contactId = currentCard.getAttribute('data-id');
        showNextCard(currentCard);

        fetch(`/api/v1/contacts/${contactId}`, {
            method: 'DELETE'
        }).catch(err => console.error('Failed to delete:', err));
    }

    window.showNextCard = function(currentCard) {
        const nextCard = currentCard.nextElementSibling;
        currentCard.remove(); // Remove from DOM

        // Decrement the UI counter
        window.decrementBadge();

        if (nextCard) {
            nextCard.classList.remove('hidden');
        } else {
            document.getElementById('contact-deck').innerHTML = `
                <div class="text-center py-20 animate-bounce">
                    <h3 class="text-xl font-bold text-success">Done!</h3>
                    <p>You've cleared the queue!</p>
                    <p>Refresh for more</p>
                </div>
                <div class="text-center bg-base-200 rounded-box">
                    <a href="/" class="btn btn-outline btn-sm mt-4">Return to Contacts</a>
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