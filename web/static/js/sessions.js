/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

// KindredCard - Settings Page JavaScript
(function() {
    'use strict';

    window.revokeSession = function(sessionId) {
        if (!confirm('Are you sure you want to revoke this session? This will log out that device.')) {
            return;
        }
        
        fetch('/api/v1/sessions/' + sessionId, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
            }
        })
        .then(response => {
            if (response.ok) {
                // Reload page to show updated sessions
                window.location.reload();
            } else {
                alert('Failed to revoke session');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Failed to revoke session');
        });
    }

    window.revokeAllOtherSessions = function() {
        if (!confirm('Are you sure you want to revoke all other sessions? This will log out all other devices.')) {
            return;
        }
        
        fetch('/api/v1/sessions/revoke-others', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            }
        })
        .then(response => {
            if (response.ok) {
                // Reload page to show updated sessions
                window.location.reload();
            } else {
                alert('Failed to revoke sessions');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Failed to revoke sessions');
        });
    }
})();