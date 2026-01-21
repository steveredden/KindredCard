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
    
    window.applyRel = async function(e, targetId, sourceId, typeId) {
        const card = e.currentTarget.closest('.util-card');
        UtilCommon.showNext(card);
        
        fetch(`/api/v1/contacts/${sourceId}/relationships`, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                related_contact_id: parseInt(targetId),
                relationship_type_id: parseInt(typeId)
            })
        });
    };

})();