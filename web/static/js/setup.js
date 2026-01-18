/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

// KindredCard - Setup Page JavaScript
(function() {
    'use strict';

    // Password validation
    const form = document.querySelector('form');
    if (form) {
        form.addEventListener('submit', function(e) {
            const password = document.querySelector('input[name="password"]').value;
            const confirm = document.querySelector('input[name="confirm_password"]').value;
            
            if (password !== confirm) {
                e.preventDefault();
                alert('Passwords do not match!');
                return false;
            }

            // Additional validation
            if (password.length < 8) {
                e.preventDefault();
                alert('Password must be at least 8 characters long!');
                return false;
            }
        });
    }

    // Real-time password match indicator
    const confirmInput = document.querySelector('input[name="confirm_password"]');
    const passwordInput = document.querySelector('input[name="password"]');
    
    if (confirmInput && passwordInput) {
        confirmInput.addEventListener('input', function() {
            if (this.value && passwordInput.value) {
                if (this.value === passwordInput.value) {
                    this.classList.remove('input-error');
                    this.classList.add('input-success');
                } else {
                    this.classList.remove('input-success');
                    this.classList.add('input-error');
                }
            } else {
                this.classList.remove('input-success', 'input-error');
            }
        });
    }

    console.log('Setup page initialized');
})();
