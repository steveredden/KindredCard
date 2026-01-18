/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

// KindredCard - Login Page JavaScript
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

        if (nextCard) {
            nextCard.classList.remove('hidden');
        } else {
            document.getElementById('contact-deck').innerHTML = `
                <div class="text-center py-20 animate-bounce">
                    <h3 class="text-xl font-bold text-success">Done!</h3>
                    <p>You've cleared the queue.</p>
                </div>`;
        }
    }
})();