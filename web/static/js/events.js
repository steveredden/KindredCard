/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

// KindredCard - Events Page

document.addEventListener("DOMContentLoaded", function() {
    // 1. Get today's date in the user's local time
    const today = new Date();
    today.setHours(0, 0, 0, 0); // Normalize to midnight

    document.querySelectorAll('[data-event-date]').forEach(el => {
        const rawDate = el.getAttribute('data-event-date');
        if (!rawDate) return;

        // 2. Parse the event date (YYYY-MM-DD)
        // Adding 'T00:00:00' ensures it's treated as local time midnight
        const eventDate = new Date(rawDate + 'T00:00:00');
        
        // 3. Calculate the literal difference in days
        const diffInMs = eventDate.getTime() - today.getTime();
        
        // Use Math.round to handle potential DST (Daylight Savings) offsets 
        // where a day might be 23 or 25 hours long.
        const diffInDays = Math.round(diffInMs / (1000 * 60 * 60 * 24));

        let tipText = "";
        if (diffInDays === 0) {
            tipText = "Today!";
        } else if (diffInDays === 1) {
            tipText = "Tomorrow";
        } else if (diffInDays === -1) {
            tipText = "Yesterday";
        } else if (diffInDays < 0) {
            tipText = `${Math.abs(diffInDays)} days ago`;
        } else {
            tipText = `in ${diffInDays} days`;
        }

        el.setAttribute('data-tip', tipText);
    });
});