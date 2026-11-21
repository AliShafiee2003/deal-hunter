// ==UserScript==
// @name         Fast Buy 1-Click for SnappShop & Digikala
// @namespace    https://snappshop.ir/
// @version      2.0
// @description  افزودن خودکار به سبد و انتقال فوری به صفحه انتخاب آدرس و شیپینگ (shipping) در اسنپ‌شاپ و دیجی‌کالا
// @match        https://snappshop.ir/*
// @match        https://www.snappshop.ir/*
// @match        https://digikala.com/*
// @match        https://www.digikala.com/*
// @run-at       document-idle
// @grant        none
// ==/UserScript==

(function() {
    'use strict';

    const isDigikala = window.location.hostname.includes('digikala.com');
    const isSnapp = window.location.hostname.includes('snappshop.ir');

    const urlParams = new URLSearchParams(window.location.search);
    const isQuickBuy = urlParams.get('quick_buy') === '1' || sessionStorage.getItem('snapp_quick_buy') === '1';

    // Helper to find button by text or selectors
    function findButtonByText(textKeywords) {
        const buttons = Array.from(document.querySelectorAll('button, a[role="button"]'));
        for (const btn of buttons) {
            const txt = (btn.textContent || '').trim();
            for (const kw of textKeywords) {
                if (txt.includes(kw) && !btn.disabled) {
                    return btn;
                }
            }
        }
        return null;
    }

    // Banner alert in page
    function showBanner(message, storeColor = '#ff3366') {
        let banner = document.getElementById('fast-buy-banner');
        if (!banner) {
            banner = document.createElement('div');
            banner.id = 'fast-buy-banner';
            banner.style.cssText = `position:fixed;top:0;left:0;right:0;background:${storeColor};color:#fff;text-align:center;padding:12px;font-size:16px;font-weight:bold;z-index:9999999;box-shadow:0 4px 20px rgba(0,0,0,0.6);direction:rtl;font-family:system-ui,sans-serif;`;
            document.body.appendChild(banner);
        }
        banner.innerHTML = '⚡ ' + message;
    }

    const themeColor = isDigikala ? 'linear-gradient(90deg, #ef394e, #c0142b)' : 'linear-gradient(90deg, #7c3aed, #ec4899)';
    const storeLabel = isDigikala ? 'دیجی‌کالا' : 'اسنپ‌شاپ';

    // 1. PRODUCT DETAIL PAGE
    if (window.location.pathname.startsWith('/product/') && isQuickBuy) {
        sessionStorage.setItem('snapp_quick_buy', '1');
        showBanner(`حالت خرید سریع ${storeLabel} فعال است: در حال افزودن خودکار به سبد خرید...`, themeColor);

        let attempts = 0;
        const interval = setInterval(() => {
            attempts++;
            const addBtn = findButtonByText(['افزودن به سبد خرید', 'افزودن به سبد']) ||
                           document.querySelector('[data-cro-id="pdp-add-to-cart"]');
            if (addBtn) {
                clearInterval(interval);
                showBanner(`کالا به سبد ${storeLabel} اضافه شد! در حال انتقال به مرحله ثبت سفارش...`, themeColor);
                addBtn.click();

                // Redirect to cart after click
                setTimeout(() => {
                    const cartURL = isDigikala ? 'https://www.digikala.com/checkout/cart?auto_order=1'
                                               : 'https://snappshop.ir/checkout/cart?auto_order=1';
                    window.location.href = cartURL;
                }, 600);
            }

            if (attempts > 50) { // 5 seconds timeout
                clearInterval(interval);
            }
        }, 100);
    }

    // 2. CHECKOUT CART PAGE
    if (window.location.pathname === '/checkout/cart' && (urlParams.get('auto_order') === '1' || sessionStorage.getItem('snapp_quick_buy') === '1')) {
        showBanner(`در حال ثبت سفارش در ${storeLabel} و انتقال مستقیم به انتخاب آدرس و شیپینگ (Shipping)...`, themeColor);

        let attempts = 0;
        const interval = setInterval(() => {
            attempts++;
            const orderBtn = findButtonByText(['ثبت سفارش', 'ادامه خرید', 'انتخاب آدرس', 'ادامه جهت تسویه حساب', 'ادامه ثبت سفارش']);
            if (orderBtn) {
                clearInterval(interval);
                sessionStorage.removeItem('snapp_quick_buy');
                orderBtn.click();

                // If click doesn't redirect immediately, navigate to shipping
                setTimeout(() => {
                    if (!window.location.pathname.includes('/shipping')) {
                        const shipURL = isDigikala ? 'https://www.digikala.com/checkout/shipping'
                                                   : 'https://snappshop.ir/checkout/shipping';
                        window.location.href = shipURL;
                    }
                }, 900);
            }

            if (attempts > 60) {
                clearInterval(interval);
                const shipURL = isDigikala ? 'https://www.digikala.com/checkout/shipping'
                                           : 'https://snappshop.ir/checkout/shipping';
                window.location.href = shipURL;
            }
        }, 100);
    }

    // 3. SHIPPING PAGE REACHED
    if (window.location.pathname.includes('/checkout/shipping')) {
        sessionStorage.removeItem('snapp_quick_buy');
        showBanner(`🎉 سفارش در ${storeLabel} ثبت شد و به صفحه انتخاب آدرس و زمان تحویل رسیدید! لطفا فاکتور را تایید نمایید.`, themeColor);
    }
})();
