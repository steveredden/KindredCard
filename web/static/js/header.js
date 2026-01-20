/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

// KindredCard - Header
(function() {
    'use strict';

    window.toggleClearButton = function(input) {
        const clearBtn = document.getElementById('clear-search');
        if (input.value.length > 0) {
            clearBtn.classList.remove('hidden');
        } else {
            clearBtn.classList.add('hidden');
            document.getElementById('search-results').innerHTML = ''; // Hide dropdown
        }
    }

    window.clearSearchInput = function() {
        const input = document.getElementById('search-input');
        input.value = '';
        toggleClearButton(input);
        document.getElementById('search-results').innerHTML = '';
        input.focus();
    }

})();