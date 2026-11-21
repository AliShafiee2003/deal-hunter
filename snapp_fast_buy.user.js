// ==UserScript==
// @name         Fast Buy 1-Click for SnappShop (Complete Auto-Checkout to Bank Gateway)
// @namespace    https://snappshop.ir/
// @version      3.0
// @description  خرید خودکار ۱-کلیکه از صفحه محصول تا درگاه بانکی: افزودن به سبد، ثبت سفارش، انتخاب خودکار اولین روز و ساعت ارسال، تایید و انتقال به درگاه بانک
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
    if (isQuickBuy) {
        sessionStorage.setItem('snapp_quick_buy', '1');
    }

    // Audio chime using Web Audio API
    function playAlertSound(freq = 1100) {
        try {
            const ctx = new (window.AudioContext || window.webkitAudioContext)();
            const osc = ctx.createOscillator();
            const gain = ctx.createGain();
            osc.type = 'triangle';
            osc.frequency.setValueAtTime(freq, ctx.currentTime);
            osc.frequency.exponentialRampToValueAtTime(freq * 1.5, ctx.currentTime + 0.25);
            gain.gain.setValueAtTime(0.3, ctx.currentTime);
            gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.4);
            osc.connect(gain);
            gain.connect(ctx.destination);
            osc.start();
            osc.stop(ctx.currentTime + 0.4);
        } catch(e) {}
    }

    // Helper to find button by text or selectors
    function findButtonByText(textKeywords) {
        const buttons = Array.from(document.querySelectorAll('button, a[role="button"], div[role="button"]'));
        for (const btn of buttons) {
            const txt = (btn.textContent || '').trim();
            for (const kw of textKeywords) {
                if (txt.includes(kw) && !btn.disabled && !btn.classList.contains('disabled')) {
                    return btn;
                }
            }
        }
        return null;
    }

    // Top Banner alert in page
    function showBanner(message, bgColor = 'linear-gradient(90deg, #dc2626, #991b1b)') {
        let banner = document.getElementById('fast-buy-banner');
        if (!banner) {
            banner = document.createElement('div');
            banner.id = 'fast-buy-banner';
            banner.style.cssText = `position:fixed;top:0;left:0;right:0;background:${bgColor};color:#fff;text-align:center;padding:12px;font-size:15px;font-weight:bold;z-index:99999999;box-shadow:0 4px 25px rgba(0,0,0,0.7);direction:rtl;font-family:system-ui,tahoma,sans-serif;`;
            if (document.body) {
                document.body.appendChild(banner);
            } else {
                document.addEventListener('DOMContentLoaded', () => document.body.appendChild(banner));
            }
        }
        banner.innerHTML = (isS26Sniper ? '🎯 [اسنایپر Galaxy S26] ' : '⚡ ') + message;
    }

    const themeColor = isS26Sniper ? 'linear-gradient(90deg, #dc2626, #b91c1c)' : 'linear-gradient(90deg, #7c3aed, #ec4899)';

    // =========================================================================
    // مرحله ۱: صفحه جزئیات کالا (PRODUCT DETAIL PAGE) -> افزودن به سبد خرید
    // =========================================================================
    if (window.location.pathname.startsWith('/product/') && (isQuickBuy || isS26Sniper)) {
        playAlertSound(880);
        showBanner(`مرحله ۱/۴: در حال افزودن خودکار به سبد خرید اسنپ‌شاپ...`, themeColor);

        let clicked = false;
        function tryAddToCart() {
            if (clicked) return;
            const addBtn = findButtonByText(['افزودن به سبد خرید', 'افزودن به سبد']) ||
                           document.querySelector('[data-cro-id="pdp-add-to-cart"]');
            if (addBtn) {
                clicked = true;
                playAlertSound(1320);
                showBanner(`✅ کالا به سبد اضافه شد! در حال انتقال به مرحله ثبت سبد...`, themeColor);
                addBtn.click();

                setTimeout(() => {
                    window.location.href = 'https://snappshop.ir/checkout/cart?auto_order=1';
                }, 350);
            }
        }

        const addInterval = setInterval(tryAddToCart, 40);
        const addObserver = new MutationObserver(tryAddToCart);
        addObserver.observe(document.documentElement, { childList: true, subtree: true });

        setTimeout(() => {
            clearInterval(addInterval);
            addObserver.disconnect();
        }, 12000);
    }

    // =========================================================================
    // مرحله ۲: صفحه سبد خرید (CART PAGE) -> کلیک ثبت سفارش
    // =========================================================================
    if (window.location.pathname === '/checkout/cart' && (isQuickBuy || isS26Sniper)) {
        playAlertSound(980);
        showBanner(`مرحله ۲/۴: در حال ثبت سفارش و انتقال به صفحه انتخاب زمان ارسال (Shipping)...`, themeColor);

        let orderClicked = false;
        function trySubmitOrder() {
            if (orderClicked) return;
            const orderBtn = findButtonByText(['ثبت سفارش', 'ادامه خرید', 'انتخاب آدرس', 'ادامه جهت تسویه حساب']);
            if (orderBtn) {
                orderClicked = true;
                playAlertSound(1400);
                showBanner(`✅ سفارش ثبت شد! در حال انتقال به انتخاب روز و ساعت ارسال...`, themeColor);
                orderBtn.click();

                setTimeout(() => {
                    if (!window.location.pathname.includes('/shipping')) {
                        window.location.href = 'https://snappshop.ir/checkout/shipping?auto_ship=1';
                    }
                }, 500);
            }
        }

        const cartInterval = setInterval(trySubmitOrder, 40);
        const cartObserver = new MutationObserver(trySubmitOrder);
        cartObserver.observe(document.documentElement, { childList: true, subtree: true });

        setTimeout(() => {
            clearInterval(cartInterval);
            cartObserver.disconnect();
        }, 10000);
    }

    // =========================================================================
    // مرحله ۳: صفحه انتخاب زمان ارسال (SHIPPING PAGE) -> انتخاب اولین روز + اولین ساعت + تایید و ادامه
    // =========================================================================
    if (window.location.pathname.includes('/checkout/shipping') && (isQuickBuy || isS26Sniper)) {
        playAlertSound(1100);
        showBanner(`مرحله ۳/۴ (Shipping): در حال انتخاب اولین روز و ساعت ارسال در دسترس...`, themeColor);

        let dayPicked = false;
        let timePicked = false;
        let continueClicked = false;

        function runShippingSequence() {
            if (continueClicked) return;

            // ۳.۱: انتخاب اولین روز فعال
            if (!dayPicked) {
                const shipmentBox = document.querySelector('[id^="shipment-"]') || document;
                const dayCards = Array.from(shipmentBox.querySelectorAll('div, button')).filter(el => {
                    const txt = (el.textContent || '').trim();
                    const hasDay = /دوشنبه|سه شنبه|چهارشنبه|پنجشنبه|جمعه|شنبه|یکشنبه/.test(txt);
                    return hasDay && txt.length < 50 && el.children.length <= 4 && !el.classList.contains('disabled') && !el.closest('.disabled');
                });

                if (dayCards.length > 0) {
                    dayCards[0].click();
                    dayPicked = true;
                }
            }

            // ۳.۲: انتخاب اولین ساعت فعال (رادیو باتن یا بازه ساعتی مانند ۱۱ الی ۱۴)
            if (dayPicked && !timePicked) {
                // جستجوی رادیوباتن‌های فعال
                const radios = Array.from(document.querySelectorAll('input[type="radio"]:not(:disabled)'));
                if (radios.length > 0) {
                    radios[0].click();
                    if (!radios[0].checked) {
                        radios[0].checked = true;
                        radios[0].dispatchEvent(new Event('change', { bubbles: true }));
                    }
                    timePicked = true;
                } else {
                    // یا کارت‌های بازه ساعتی
                    const slotDivs = Array.from(document.querySelectorAll('div, label, span')).filter(el => {
                        const txt = (el.textContent || '').trim();
                        return /\d+\s*الی\s*\d+/.test(txt) && txt.length < 40 && el.children.length <= 3 && !el.closest('.disabled');
                    });
                    if (slotDivs.length > 0) {
                        slotDivs[0].click();
                        timePicked = true;
                    }
                }
            }

            // ۳.۳: کلیک روی دکمه "تایید و ادامه" پس از مشخص شدن زمان
            if (timePicked) {
                const continueBtn = findButtonByText(['تایید و ادامه', 'ادامه و تایید', 'ادامه جهت پرداخت']);
                if (continueBtn && !continueBtn.disabled && !continueBtn.classList.contains('disabled')) {
                    continueClicked = true;
                    playAlertSound(1500);
                    showBanner(`✅ زمان ارسال تایید شد! در حال انتقال به صفحه پرداخت...`, themeColor);
                    continueBtn.click();

                    setTimeout(() => {
                        if (!window.location.pathname.includes('/payment')) {
                            window.location.href = 'https://snappshop.ir/checkout/payment?auto_pay=1';
                        }
                    }, 800);
                }
            }
        }

        const shipInterval = setInterval(runShippingSequence, 60);
        const shipObserver = new MutationObserver(runShippingSequence);
        shipObserver.observe(document.documentElement, { childList: true, subtree: true });

        setTimeout(() => {
            clearInterval(shipInterval);
            shipObserver.disconnect();
        }, 15000);
    }

    // =========================================================================
    // مرحله ۴: صفحه پرداخت (PAYMENT PAGE) -> کلیک "پرداخت از طریق درگاه بانکی"
    // =========================================================================
    if (window.location.pathname.includes('/checkout/payment') && (isQuickBuy || isS26Sniper)) {
        playAlertSound(1200);
        showBanner(`مرحله ۴/۴ (پرداخت نهایی): در حال کلیک روی پرداخت از طریق درگاه بانکی...`, themeColor);

        let payClicked = false;
        function tryPayment() {
            if (payClicked) return;
            const payBtn = findButtonByText([
                'پرداخت از طریق درگاه بانکی',
                'پرداخت از طریق درگاه',
                'پرداخت اینترنتی',
                'پرداخت آنلاین',
                'پرداخت'
            ]);

            if (payBtn && !payBtn.disabled && !payBtn.classList.contains('disabled')) {
                payClicked = true;
                sessionStorage.removeItem('snapp_quick_buy');
                sessionStorage.removeItem('snapp_s26_sniper');
                playAlertSound(1760);
                showBanner(`🚀 انتقال مستقیم به درگاه بانکی شاپرک انجام شد! لطفا رمز پویا و اطلاعات کارت را وارد فرمایید.`, 'linear-gradient(90deg, #16a34a, #15803d)');
                payBtn.click();
            }
        }

        const payInterval = setInterval(tryPayment, 50);
        const payObserver = new MutationObserver(tryPayment);
        payObserver.observe(document.documentElement, { childList: true, subtree: true });

        setTimeout(() => {
            clearInterval(payInterval);
            payObserver.disconnect();
        }, 15000);
    }
})();
