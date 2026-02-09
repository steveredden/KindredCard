/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

// KindredCard - Utilities page: Address Patching
(function() {
    'use strict';
    
    window.syncAddress = async function(e, targetId, addressObj) {
        const btn = e.currentTarget;
        const card = btn.closest('.util-card');
        btn.disabled = true;

        let body = {
            contact_id: parseInt(targetId, 10),
            street: addressObj.street,
            extended_street: addressObj.extended_street,
            city: addressObj.city,
            state: addressObj.state,
            postal_code: addressObj.postal_code,
            country: addressObj.country,
            label_type_id: parseInt(addressObj.label_type_id, 10),
            is_primary: addressObj.is_primary || false
        };

        const res = await fetch(`/api/v1/contacts/${targetId}/addresses`, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(body)
        });

        if (res.ok) UtilCommon.showNext(card);
        else btn.disabled = false;
    };
    
})();