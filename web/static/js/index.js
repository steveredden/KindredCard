/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

// KindredCard - Index Page JavaScript (Simplified)
(function() {
    'use strict';

    let currentContactId = null;
    let isCardFlipped = false;

    // Filter contacts by search
    window.filterContacts = function(query) {
        const cards = document.querySelectorAll('.contact-card');
        const lowerQuery = query.toLowerCase();
        
        cards.forEach(card => {
            const name = (card.dataset.contactName || '').toLowerCase();
            const nickname = (card.dataset.contactNickname || '').toLowerCase();
            const maiden = (card.dataset.contactMaiden || '').toLowerCase();
            const email = (card.dataset.contactEmail || '').toLowerCase();
            const phone = (card.dataset.contactPhone || '').toLowerCase();
            
            const matches = name.includes(lowerQuery) ||
                            nickname.includes(lowerQuery) ||
                            maiden.includes(lowerQuery) ||
                            email.includes(lowerQuery) ||
                            phone.includes(lowerQuery);
        
            card.style.display = matches ? '' : 'none';
        });
    };

    // Open contact card modal
    window.openContactCard = function(contactId) {
        currentContactId = contactId;
        isCardFlipped = false;
        
        // Find the contact card that was clicked
        const contactCard = document.querySelector(`[data-contact-id="${contactId}"]`);
        if (!contactCard) {
            showNotification('Contact not found', 'error');
            return;
        }

        // Get the name from the card
        const name = contactCard.dataset.contactName;
        document.getElementById('modalContactName').textContent = name;

        // Reset flip
        document.getElementById('cardFront').style.transform = 'rotateY(0deg)';
        document.getElementById('cardBack').style.transform = 'rotateY(180deg)';

        // Load content from templates
        const frontTemplate = contactCard.querySelector('.modal-front-content');
        const backTemplate = contactCard.querySelector('.modal-back-content');
        
        if (frontTemplate) {
            document.getElementById('cardFrontContent').innerHTML = frontTemplate.innerHTML;
        }
        
        if (backTemplate) {
            document.getElementById('cardBackContent').innerHTML = backTemplate.innerHTML;
        }

        // Open modal
        openModal('contactCardModal');
    };

    // Flip card animation (now callable by clicking anywhere on card)
    window.flipCard = function(event) {
        // Prevent flip if clicking on action buttons
        if (event && event.target.closest('.modal-action')) {
            return;
        }
        
        const front = document.getElementById('cardFront');
        const back = document.getElementById('cardBack');
        
        if (isCardFlipped) {
            front.style.transform = 'rotateY(0deg)';
            back.style.transform = 'rotateY(180deg)';
        } else {
            front.style.transform = 'rotateY(-180deg)';
            back.style.transform = 'rotateY(0deg)';
        }
        
        isCardFlipped = !isCardFlipped;
    };

    // Edit contact
    window.editContact = function(event) {
        event?.stopPropagation(); // Prevent card flip
        if (currentContactId) {
            window.location.href = `/contacts/${currentContactId}`;
        }
    };

    // Export contact vCard
    window.exportContactVCard = async function(event) {
        event?.stopPropagation(); // Prevent card flip
        if (!currentContactId) return;
        
        // Get contact name from modal
        const name = document.getElementById('modalContactName').textContent;
        
        try {
            const response = await fetch(`/api/v1/contacts/${currentContactId}/vcard`);
            if (!response.ok) throw new Error('Export failed');
            
            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `${name}.vcf`;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            a.remove();
            
            showNotification('Contact exported successfully', 'success');
        } catch (error) {
            console.error('Export error:', error);
            showNotification('Failed to export contact', 'error');
        }
    };

    // Delete contact with modal
    window.openDeleteModal = function() {
        const modal = document.getElementById('deleteModal');
        if (modal) {
            modal.showModal();
        }
    };

    window.closeDeleteModal = function() {
        const modal = document.getElementById('deleteModal');
        if (modal) {
            modal.close();
        }
    };

    window.confirmDelete = async function() {
        const contactId = currentContactId
        
        try {
            const response = await fetch(`/api/v1/contacts/${contactId}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                showNotification('Contact deleted successfully', 'success');
                closeDeleteModal();
                setTimeout(() => window.location.href = '/', 800);
            } else {
                throw new Error('Delete failed');
            }
        } catch (error) {
            console.error('Delete error:', error);
            showNotification('Failed to delete contact', 'error');
        }
    };

    // Add contact modal
    window.openAddContactModal = function() {
        document.getElementById('addContactForm')?.reset();
        openModal('addContactModal');
    };

    // Handle add contact form submission
    const addContactForm = document.getElementById('addContactForm');
    if (addContactForm) {
        addContactForm.addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const formData = new FormData(this);
            const contact = {
                given_name: formData.get('given_name'),
                family_name: formData.get('family_name'),
                gender: formData.get('gender')
            };
            
            try {
                const response = await fetch('/api/v1/contacts', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(contact)
                });
                
                if (response.ok) {
                    showNotification('Contact created successfully', 'success');
                    closeModal('addContactModal');
                    setTimeout(() => window.location.reload(), 500);
                } else {
                    throw new Error('Failed to create contact');
                }
            } catch (error) {
                console.error('Create error:', error);
                showNotification('Failed to create contact', 'error');
            }
        });
    }

    window.sortContacts = function(criteria) {
        const container = document.getElementById('contactsGallery');
        const cards = Array.from(container.getElementsByClassName('contact-card'));

        cards.sort((a, b) => {
            // Helper to safely get attributes
            const getAttr = (el, attr) => (el.getAttribute(attr) || "").toLowerCase();
            const getId = (el) => parseInt(el.getAttribute('data-contact-id') || 0);

            switch (criteria) {
                case 'lastname-asc': {
                    const lastA = getAttr(a, 'data-contact-lastname');
                    const lastB = getAttr(b, 'data-contact-lastname');
                    
                    // 1. Primary Sort: Last Name
                    const primarySort = lastA.localeCompare(lastB);
                    
                    // 2. Secondary Sort: If last names are equal, sort by First Name
                    if (primarySort === 0) {
                        const firstA = getAttr(a, 'data-contact-name');
                        const firstB = getAttr(b, 'data-contact-name');
                        return firstA.localeCompare(firstB);
                    }
                    
                    return primarySort;
                }

                case 'name-asc': {
                    const nameA = getAttr(a, 'data-contact-name');
                    const nameB = getAttr(b, 'data-contact-name');
                    return nameA.localeCompare(nameB);
                }

                case 'id-desc':
                    return getId(b) - getId(a);

                case 'id-asc':
                    return getId(a) - getId(b);

                default:
                    return 0;
            }
        });

        // Re-append cards in the new order
        cards.forEach(card => container.appendChild(card));
    };

    console.log('Index page initialized (simplified)');
})();
