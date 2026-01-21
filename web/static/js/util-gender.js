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
        const card = document.querySelector('.util-card:not(.hidden)');
        if (!card) return;
        const id = card.getAttribute('data-id');
        
        UtilCommon.showNext(card);
        fetch(`/api/v1/contacts/${id}`, {
            method: 'PATCH',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({gender: gender})
        });
    };
})();