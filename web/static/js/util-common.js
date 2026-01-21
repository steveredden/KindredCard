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
    
    window.UtilCommon = {
        // Standardized way to move to the next item
        showNext: function(currentCard) {
            const nextCard = currentCard.nextElementSibling;
            
            // 1. Animate out
            currentCard.classList.add('opacity-0', 'scale-95', 'pointer-events-none');
            
            setTimeout(() => {
                currentCard.remove();
                this.updateBadge();

                if (nextCard && nextCard.classList.contains('util-card')) {
                    nextCard.classList.remove('hidden');
                    // Trigger any per-page logic (like phone preview)
                    if (window.onCardChange) window.onCardChange(nextCard);
                } else {
                    this.showFinishedState();
                }
            }, 200);
        },

        updateBadge: function() {
            const badge = document.getElementById('remaining-count');
            const count = document.querySelectorAll('.util-card').length;
            if (badge) badge.innerText = `${count} remaining`;
        },

        showFinishedState: function() {
            // Hide side buttons (F/M, etc)
            document.querySelectorAll('.side-action-btn').forEach(el => el.classList.add('hidden'));
            
            const deck = document.getElementById('contact-deck');
            deck.innerHTML = `
                <div class="text-center py-16 bg-base-200 rounded-3xl border-2 border-dashed border-base-300 animate-in fade-in zoom-in duration-300">
                    <div class="text-6xl mb-4">🎉</div>
                    <h3 class="text-2xl font-bold">All Caught Up!</h3>
                    <p class="opacity-60 mb-8">This utility has no more pending items.</p>
                    <div class="flex justify-center gap-4">
                        <a href="/" class="btn btn-primary px-8">Return Home</a>
                        <a href="/settings" class="btn btn-ghost">Settings</a>
                    </div>
                </div>`;
        }
    };

})();