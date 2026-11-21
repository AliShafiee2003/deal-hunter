// ==UserScript==
// @name         Fast Buy 1-Click for SnappShop & Digikala (with Galaxy S26 Sniper)
// @namespace    https://snappshop.ir/
// @version      2.5
// @description  افزودن خودکار به سبد و انتقال فوری به صفحه انتخاب آدرس و شیپینگ (shipping) در اسنپ‌شاپ و دیجی‌کالا + اسنایپر خودکار Galaxy S26
// @match        https://snappshop.ir/*
// @match        https://www.snappshop.ir/*
// @match        https://digikala.com/*
// @match        https://www.digikala.com/*
// @run-at       document-start
// @grant        none
// ==/UserScript==

(function() {
    'use strict';

    const isDigikala = window.location.hostname.includes('digikala.com');
    const isSnapp = window.location.hostname.includes('snappshop.ir');

    const urlParams = new URLSearchParams(window.location.search);
    const isQuickBuy = urlParams.get('quick_buy') === '1' || sessionStorage.getItem('snapp_quick_buy') === '1';
    const isS26Sniper = urlParams.get('sniper') === 's26' || sessionStorage.getItem('snapp_s26_sniper') === '1';

    if (isS26Sniper) {
        sessionStorage.setItem('snapp_s26_sniper', '1');
    }

    // Sound alert using Web Audio API
    function playAlertSound() {
        try {
            const ctx = new (window.AudioContext || window.webkitAudioContext)();
            const osc = ctx.createOscillator();
            const gain = ctx.createGain();
            osc.type = 'triangle';
            osc.frequency.setValueAtTime(880, ctx.currentTime);
            osc.frequency.exponentialRampToValueAtTime(1760, ctx.currentTime + 0.3);
            gain.gain.setValueAtTime(0.3, ctx.currentTime);
            gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.5);
            osc.connect(gain);
            gain.connect(ctx.destination);
            osc.start();
            osc.stop(ctx.currentTime + 0.5);
        } catch(e) {}
    }

    // Helper to find button by text or selectors
    function findButtonByText(textKeywords) {
        const buttons = Array.from(document.querySelectorAll('button, a[role="button"], div[role="button"]'));
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
            banner.style.cssText = `position:fixed;top:0;left:0;right:0;background:${storeColor};color:#fff;text-align:center;padding:14px;font-size:16px;font-weight:bold;z-index:9999999;box-shadow:0 4px 25px rgba(0,0,0,0.7);direction:rtl;font-family:system-ui,tahoma,sans-serif;`;
            if (document.body) {
                document.body.appendChild(banner);
            } else {
                document.addEventListener('DOMContentLoaded', () => document.body.appendChild(banner));
            }
        }
        banner.innerHTML = (isS26Sniper ? '🎯 [اسنایپر Galaxy S26] ' : '⚡ ') + message;
    }

    const themeColor = isS26Sniper ? 'linear-gradient(90deg, #dc2626, #b91c1c)' : (isDigikala ? 'linear-gradient(90deg, #ef394e, #c0142b)' : 'linear-gradient(90deg, #7c3aed, #ec4899)');
    const storeLabel = isDigikala ? 'دیجی‌کالا' : 'اسنپ‌شاپ';

    // 1. PRODUCT DETAIL PAGE
    if (window.location.pathname.startsWith('/product/') && (isQuickBuy || isS26Sniper)) {
        sessionStorage.setItem('snapp_quick_buy', '1');
        playAlertSound();
        showBanner(`خرید سریع ${storeLabel} فعال شد! در حال افزودن خودکار به سبد خرید...`, themeColor);

        let clicked = false;
        function tryAddToCart() {
            if (clicked) return;
            const addBtn = findButtonByText(['افزودن به سبد خرید', 'افزودن به سبد']) ||
                           document.querySelector('[data-cro-id="pdp-add-to-cart"]');
            if (addBtn) {
                clicked = true;
                playAlertSound();
                showBanner(`✅ کالا به سبد ${storeLabel} اضافه شد! در حال هدایت مستقیم به ثبت نهایی و انتخاب آدرس...`, themeColor);
                addBtn.click();

                // Redirect to cart after click
                setTimeout(() => {
                    const cartURL = isDigikala ? 'https://www.digikala.com/checkout/cart?auto_order=1'
                                               : 'https://snappshop.ir/checkout/cart?auto_order=1';
                    window.location.href = cartURL;
                }, 400);
            }
        }

        // Fast polling + MutationObserver for instant detection
        const fastInterval = setInterval(() => {
            tryAddToCart();
            if (clicked) clearInterval(fastInterval);
        }, 40);

        const observer = new MutationObserver(() => {
            tryAddToCart();
            if (clicked) observer.disconnect();
        });
        observer.observe(document.documentElement, { childList: true, subtree: true });

        // Safety timeout
        setTimeout(() => {
            clearInterval(fastInterval);
            observer.disconnect();
        }, 10000);
    }

    // 2. CHECKOUT CART PAGE
    if (window.location.pathname === '/checkout/cart' && (urlParams.get('auto_order') === '1' || sessionStorage.getItem('snapp_quick_buy') === '1')) {
        playAlertSound();
        showBanner(`در حال ثبت سفارش در ${storeLabel} و انتقال مستقیم به انتخاب آدرس و شیپینگ (Shipping)...`, themeColor);

        let orderClicked = false;
        function trySubmitOrder() {
            if (orderClicked) return;
            const orderBtn = findButtonByText(['ثبت سفارش', 'ادامه خرید', 'انتخاب آدرس', 'ادامه جهت تسویه حساب', 'ادامه ثبت سفارش']);
            if (orderBtn) {
                orderClicked = true;
                sessionStorage.removeItem('snapp_quick_buy');
                orderBtn.click();

                setTimeout(() => {
                    if (!window.location.pathname.includes('/shipping')) {
                        const shipURL = isDigikala ? 'https://www.digikala.com/checkout/shipping'
                                                   : 'https://snappshop.ir/checkout/shipping';
                        window.location.href = shipURL;
                    }
                }, 600);
            }
        }

        const cartInterval = setInterval(() => {
            trySubmitOrder();
            if (orderClicked) clearInterval(cartInterval);
        }, 50);

        const cartObserver = new MutationObserver(() => {
            trySubmitOrder();
            if (orderClicked) cartObserver.disconnect();
        });
        cartObserver.observe(document.documentElement, { childList: true, subtree: true });

        setTimeout(() => {
            clearInterval(cartInterval);
            cartObserver.disconnect();
            if (!orderClicked && !window.location.pathname.includes('/shipping')) {
                const shipURL = isDigikala ? 'https://www.digikala.com/checkout/shipping'
                                           : 'https://snappshop.ir/checkout/shipping';
                window.location.href = shipURL;
            }
        }, 4000);
    }

    // 3. SHIPPING PAGE REACHED
    if (window.location.pathname.includes('/checkout/shipping')) {
        sessionStorage.removeItem('snapp_quick_buy');
        sessionStorage.removeItem('snapp_s26_sniper');
        playAlertSound();
        showBanner(`🎉 تبریک! کالا در سبد ثبت شد و به صفحه شیپینگ و آدرس تحویل رسیدید. لطفا فاکتور را تایید نمایید.`, themeColor);
    }
})();
