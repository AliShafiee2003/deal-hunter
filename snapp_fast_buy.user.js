// ==UserScript==
// @name         Fast Buy 1-Click for SnappShop (Complete Auto-Checkout to Bank Gateway)
// @namespace    https://snappshop.ir/
// @version      4.2
// @description  اسنایپر هوشمند Galaxy S26: رصد خودکار لیست کالاها و لیزی‌لود (Product-List)، افزودن به سبد، ثبت سفارش، انتخاب سریعترین ارسال و اتصال مستقیم به درگاه بانکی
// @match        https://snappshop.ir/*
// @match        https://www.snappshop.ir/*
// @run-at       document-start
// @grant        none
// ==/UserScript==

(function() {
    'use strict';

    const urlParams = new URLSearchParams(window.location.search);
    const isQuickBuy = urlParams.get('quick_buy') === '1' || sessionStorage.getItem('snapp_quick_buy') === '1';
    const isS26Sniper = urlParams.get('sniper') === 's26' || sessionStorage.getItem('snapp_s26_sniper') === '1';
    const isProductListPage = window.location.pathname.startsWith('/product-list/');

    if (isS26Sniper) sessionStorage.setItem('snapp_s26_sniper', '1');
    if (isQuickBuy)  sessionStorage.setItem('snapp_quick_buy', '1');

    // -------------------------------------------------------------------------
    // Audio Chime
    // -------------------------------------------------------------------------
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

    // -------------------------------------------------------------------------
    // Banner UI
    // -------------------------------------------------------------------------
    const themeColor = 'linear-gradient(90deg, #dc2626, #b91c1c)';

    function showBanner(message, bgColor) {
        bgColor = bgColor || themeColor;
        let banner = document.getElementById('fast-buy-banner');
        if (!banner) {
            banner = document.createElement('div');
            banner.id = 'fast-buy-banner';
            banner.style.cssText = 'position:fixed;top:0;left:0;right:0;color:#fff;text-align:center;padding:14px 10px;font-size:15px;font-weight:bold;z-index:99999999;box-shadow:0 4px 25px rgba(0,0,0,0.7);direction:rtl;font-family:system-ui,tahoma,sans-serif;';
            const attach = () => document.body && document.body.appendChild(banner);
            if (document.body) attach(); else document.addEventListener('DOMContentLoaded', attach);
        }
        banner.style.background = bgColor;
        banner.innerHTML = '🎯 [اسنایپر Galaxy S26] ' + message;
    }

    // -------------------------------------------------------------------------
    // Persian digits normalizer
    // -------------------------------------------------------------------------
    function normalizeDigits(str) {
        return (str || '')
            .replace(/[۰-۹]/g, d => '۰۱۲۳۴۵۶۷۸۹'.indexOf(d))
            .replace(/[٠-٩]/g, d => '٠١٢٣٤٥٦٧٨٩'.indexOf(d));
    }

    // -------------------------------------------------------------------------
    // S26 Matcher Logic
    // -------------------------------------------------------------------------
    function isMatchingS26(title, text) {
        const full = normalizeDigits((title + ' ' + text).toLowerCase());
        const hasS26 = /s\s*26|s26|اس\s*۲۶|اس۲۶/.test(full);
        if (!hasS26) return false;

        // Reject accessories
        const isAccessory = /قاب|کاور|گلس|محافظ|شارژر|کابل|هولدر|استند|کیف|بند|هندزفری|هدفون|باتری|آداپتور/.test(full);
        if (isAccessory) return false;

        return true;
    }

    // -------------------------------------------------------------------------
    // Element Finder
    // -------------------------------------------------------------------------
    function findByText(keywords, selectors) {
        selectors = selectors || 'button, a[role="button"], div[role="button"], [class*="btn"]';
        const els = Array.from(document.querySelectorAll(selectors));
        for (const el of els) {
            const txt = (el.textContent || '').trim();
            for (const kw of keywords) {
                if (txt.includes(kw) && !el.disabled && !el.classList.contains('disabled') && !el.closest('[disabled]')) {
                    return el;
                }
            }
        }
        return null;
    }

    // =========================================================================
    // مرحله ۰: پایش ویژه لیست محصولات و رویدادها (مانند Doe3z0) با پشتیبانی Lazy Loading
    // =========================================================================
    if (isProductListPage) {
        showBanner('پایش آنی لیست کالاها و لیزی‌لود برای Galaxy S26 فعال است...');

        let s26Found = false;
        function scanProductCards() {
            if (s26Found) return;
            const cards = Array.from(document.querySelectorAll('a[href*="/product/snp-"]'));

            for (const card of cards) {
                const text = card.textContent || '';
                const titleEl = card.querySelector('h2, h3, h4, p, [class*="title"]') || card;
                const title = titleEl.textContent || '';

                if (isMatchingS26(title, text)) {
                    // Extract price numbers
                    const normText = normalizeDigits(text).replace(/,/g, '');
                    const priceMatches = normText.match(/\d{7,10}/g);
                    let validPrice = false;

                    if (priceMatches) {
                        for (const pStr of priceMatches) {
                            let p = parseInt(pStr, 10);
                            if (p > 100000000) p = p / 10;
                            // Check 20M to 30M Toman
                            if (p >= 20000000 && p <= 30000000) {
                                validPrice = true;
                                break;
                            }
                        }
                    }

                    // If matches S26 and within 20M-30M Toman (or specifically 26M)
                    if (validPrice || /26\s*000\s*000|۲۶\s*۰۰۰\s*۰۰۰/.test(text)) {
                        s26Found = true;
                        playAlertSound(1500);
                        showBanner('🚨 گلکسی S26 شناسایی شد! انتقال فوری به فرآیند خرید...');

                        let targetHref = card.getAttribute('href') || '';
                        if (!targetHref.startsWith('http')) {
                            targetHref = 'https://snappshop.ir' + targetHref;
                        }
                        const sep = targetHref.includes('?') ? '&' : '?';
                        window.location.href = targetHref + sep + 'quick_buy=1&sniper=s26';
                        return;
                    }
                }
            }
        }

        const scanIv = setInterval(scanProductCards, 60);
        new MutationObserver(scanProductCards).observe(document.documentElement, { childList: true, subtree: true });
        window.addEventListener('scroll', scanProductCards, { passive: true });
    }

    // =========================================================================
    // مرحله ۱: صفحه کالا — افزودن به سبد خرید
    // =========================================================================
    if (window.location.pathname.startsWith('/product/') && (isQuickBuy || isS26Sniper)) {
        playAlertSound(880);
        showBanner('مرحله ۱/۴: افزودن خودکار کالا به سبد خرید...');

        let clicked = false;
        function tryAddToCart() {
            if (clicked) return;
            const btn = findByText(['افزودن به سبد خرید', 'افزودن به سبد'])
                     || document.querySelector('[data-cro-id="pdp-add-to-cart"]');
            if (btn) {
                clicked = true;
                btn.click();
                playAlertSound(1320);
                showBanner('✅ کالا به سبد اضافه شد! انتقال به مرحله ثبت سبد...');
                setTimeout(() => {
                    window.location.href = 'https://snappshop.ir/checkout/cart?auto_order=1';
                }, 350);
            }
        }

        const iv = setInterval(tryAddToCart, 40);
        new MutationObserver(tryAddToCart).observe(document.documentElement, { childList: true, subtree: true });
        setTimeout(() => clearInterval(iv), 12000);
    }

    // =========================================================================
    // مرحله ۲: صفحه سبد — کلیک ثبت سفارش
    // =========================================================================
    if (window.location.pathname === '/checkout/cart' && (isQuickBuy || isS26Sniper)) {
        playAlertSound(980);
        showBanner('مرحله ۲/۴: ثبت سفارش...');

        let done = false;
        function tryOrder() {
            if (done) return;
            const btn = findByText(['ثبت سفارش', 'ادامه خرید', 'انتخاب آدرس', 'ادامه جهت تسویه']);
            if (btn) {
                done = true;
                btn.click();
                playAlertSound(1400);
                showBanner('✅ سفارش ثبت شد! انتقال به انتخاب زمان ارسال...');
                setTimeout(() => {
                    if (!window.location.pathname.includes('/shipping')) {
                        window.location.href = 'https://snappshop.ir/checkout/shipping?auto_ship=1';
                    }
                }, 500);
            }
        }

        const iv = setInterval(tryOrder, 40);
        new MutationObserver(tryOrder).observe(document.documentElement, { childList: true, subtree: true });
        setTimeout(() => clearInterval(iv), 10000);
    }

    // =========================================================================
    // مرحله ۳: صفحه ارسال — SEQUENCE واکنشی و دقیق
    // ۱. کلیک و باز کردن بخش «سریع‌ترین ارسال»
    // ۲. کلیک اولین تاریخ در دسترس (دوشنبه، فردا و...)
    // ۳. کلیک اولین ساعت در دسترس (۱۱ الی ۱۴ و...)
    // ۴. کلیک دکمه «تایید و ادامه»
    // =========================================================================
    if (window.location.pathname.includes('/checkout/shipping') && (isQuickBuy || isS26Sniper)) {
        playAlertSound(1100);
        showBanner('مرحله ۳/۴: در حال هماهنگی زمان ارسال (Shipping)...');

        let lastFastestClick = 0;
        let daySelected = false;
        let timeSelected = false;
        let confirmClicked = false;

        function getFastestShippingElement() {
            const all = Array.from(document.querySelectorAll('div, label, section, button'));
            for (const el of all) {
                const txt = (el.textContent || '').trim();
                if ((txt.includes('سریع‌ترین ارسال') || txt.includes('سریعترین ارسال')) && txt.length < 80) {
                    return el;
                }
            }
            return null;
        }

        function getFirstDayCard() {
            const all = Array.from(document.querySelectorAll('label, div, button'));
            return all.find(el => {
                const txt = (el.textContent || '').trim();
                return /فردا|دوشنبه|سه.?شنبه|چهارشنبه|پنجشنبه|جمعه|شنبه|یکشنبه/.test(txt)
                    && txt.length < 40
                    && el.children.length <= 4
                    && !el.classList.contains('disabled')
                    && !el.closest('[class*="disabled"]');
            }) || null;
        }

        function getFirstTimeSlot() {
            const all = Array.from(document.querySelectorAll('label, div, span, button'));
            return all.find(el => {
                const txt = (el.textContent || '').trim();
                return /\d+\s*الی\s*\d+/.test(txt)
                    && txt.length < 50
                    && el.children.length <= 3
                    && !el.classList.contains('disabled')
                    && !el.closest('[class*="disabled"]');
            }) || null;
        }

        function getConfirmButton() {
            return findByText(['تایید و ادامه', 'ادامه و تایید', 'ادامه جهت پرداخت']);
        }

        function runShippingReactive() {
            if (confirmClicked) return;

            // گام ۴: کلیک دکمه تایید و ادامه
            const confirmBtn = getConfirmButton();
            if (confirmBtn && timeSelected) {
                confirmClicked = true;
                confirmBtn.click();
                playAlertSound(1500);
                showBanner('✅ زمان ارسال تایید شد! انتقال به درگاه پرداخت...');
                setTimeout(() => {
                    if (!window.location.pathname.includes('/payment')) {
                        window.location.href = 'https://snappshop.ir/checkout/payment?auto_pay=1';
                    }
                }, 800);
                return;
            }

            // گام ۳: انتخاب اولین ساعت آزاد
            const timeSlot = getFirstTimeSlot();
            if (timeSlot && !timeSelected) {
                timeSlot.click();
                const radio = timeSlot.querySelector('input[type="radio"]') ||
                              timeSlot.closest('label')?.querySelector('input[type="radio"]');
                if (radio) {
                    radio.checked = true;
                    radio.dispatchEvent(new Event('change', { bubbles: true }));
                }
                timeSelected = true;
                showBanner('مرحله ۳/۴: ساعت ارسال انتخاب شد — در انتظار دکمه تایید و ادامه...');
                return;
            }

            // گام ۲: انتخاب اولین تاریخ آزاد
            const dayCard = getFirstDayCard();
            if (dayCard && !daySelected) {
                dayCard.click();
                daySelected = true;
                showBanner('مرحله ۳/۴: روز ارسال انتخاب شد — در حال بارگذاری ساعت‌های ارسال...');
                return;
            }

            // گام ۱: باز کردن تب سریع‌ترین ارسال
            if (!dayCard && (Date.now() - lastFastestClick > 500)) {
                const fastestEl = getFastestShippingElement();
                if (fastestEl) {
                    lastFastestClick = Date.now();
                    fastestEl.click();
                    const innerRadio = fastestEl.querySelector('input[type="radio"]');
                    if (innerRadio) {
                        innerRadio.click();
                        innerRadio.checked = true;
                        innerRadio.dispatchEvent(new Event('change', { bubbles: true }));
                    }
                    showBanner('مرحله ۳/۴: تب سریع‌ترین ارسال باز شد — در حال انتخاب تاریخ...');
                }
            }
        }

        const iv = setInterval(runShippingReactive, 50);
        new MutationObserver(runShippingReactive).observe(document.documentElement, { childList: true, subtree: true });
        setTimeout(() => clearInterval(iv), 25000);
    }

    // =========================================================================
    // مرحله ۴: صفحه پرداخت — کلیک «پرداخت از طریق درگاه بانکی»
    // =========================================================================
    if (window.location.pathname.includes('/checkout/payment') && (isQuickBuy || isS26Sniper)) {
        playAlertSound(1200);
        showBanner('مرحله ۴/۴: در حال انتقال به درگاه بانکی شاپرک...');

        let done = false;
        function tryPay() {
            if (done) return;
            const btn = findByText([
                'پرداخت از طریق درگاه بانکی',
                'پرداخت از طریق درگاه',
                'پرداخت اینترنتی',
                'پرداخت آنلاین',
                'پرداخت'
            ]);
            if (btn && !btn.disabled && !btn.classList.contains('disabled')) {
                done = true;
                sessionStorage.removeItem('snapp_quick_buy');
                sessionStorage.removeItem('snapp_s26_sniper');
                btn.click();
                playAlertSound(1760);
                showBanner('🚀 در حال اتصال به درگاه بانکی شاپرک! لطفا رمز پویا را وارد فرمایید.',
                    'linear-gradient(90deg, #16a34a, #15803d)');
            }
        }

        const iv = setInterval(tryPay, 50);
        new MutationObserver(tryPay).observe(document.documentElement, { childList: true, subtree: true });
        setTimeout(() => clearInterval(iv), 15000);
    }

})();
