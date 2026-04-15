console.log("if you like superkit consider giving it a star on GitHub.");

window.trackAddToCart = function(id, name, price, currency) {
    console.log('FB Event: AddToCart', { id, name, price, currency });
    if (typeof fbq === 'function') {
        fbq('track', 'AddToCart', {
            content_ids: [id],
            content_name: name,
            content_type: 'product',
            value: price,
            currency: currency
        });
    }
};

window.trackInitiateCheckout = function(id, name, price, currency) {
    console.log('FB Event: InitiateCheckout', { id, name, price, currency });
    if (typeof fbq === 'function') {
        fbq('track', 'InitiateCheckout', {
            content_ids: [id],
            content_name: name,
            content_type: 'product',
            value: price,
            currency: currency
        });
    }
};

window.trackPurchase = function(currency, value, trackValue) {
    console.log('FB Event: Purchase', { currency, value, trackValue });
    if (typeof fbq === 'function') {
        fbq('track', 'Purchase', { 
            currency: currency, 
            value: trackValue ? value : undefined 
        });
    }
};

window.closeQuickView = function() {
    const modal = document.getElementById('quick-view-modal');
    if (modal) modal.remove();
};
