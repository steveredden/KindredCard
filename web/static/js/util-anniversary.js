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
    
    window.syncAnniversary = async function(e, targetId, val) {
        const btn = e.currentTarget;
        const card = btn.closest('.util-card');
        btn.disabled = true;

        let payload = {};
        const parts = val.split('-');
        if (parts.length === 3) payload.anniversary = val + "T00:00:00Z";
        else { payload.anniversary_month = parseInt(parts[0]); payload.anniversary_day = parseInt(parts[1]); }

        const res = await fetch(`/api/v1/contacts/${targetId}/anniversary`, {
            method: 'PATCH',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(payload)
        });

        if (res.ok) UtilCommon.showNext(card);
        else btn.disabled = false;
    };
    
})();