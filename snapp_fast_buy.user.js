// ==UserScript==
// @name         Fast Buy 1-Click for SnappShop (Complete Auto-Checkout to Bank Gateway)
// @namespace    https://snappshop.ir/
// @version      4.0
// @description  خرید خودکار ۱-کلیکه از صفحه محصول تا درگاه بانکی: افزودن به سبد، ثبت سفارش، انتخاب سریعترین ارسال، اولین روز و ساعت، تایید و انتقال به درگاه بانک
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

    if (isS26Sniper) sessionStorage.setItem('snapp_s26_sniper', '1');
    if (isQuickBuy)  sessionStorage.setItem('snapp_quick_buy', '1');

    if (!isQuickBuy && !isS26Sniper) return; // early exit — no active session

    // -------------------------------------------------------------------------
    // Audio chime
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
    // Top banner
    // -------------------------------------------------------------------------
    const themeColor = isS26Sniper
        ? 'linear-gradient(90deg, #dc2626, #b91c1c)'
        : 'linear-gradient(90deg, #7c3aed, #ec4899)';

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
        banner.innerHTML = (isS26Sniper ? '🎯 [اسنایپر Galaxy S26] ' : '⚡ ') + message;
    }

    // -------------------------------------------------------------------------
    // Helper: find clickable element by text
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

    // -------------------------------------------------------------------------
    // Helper: sleep
    // -------------------------------------------------------------------------
    function sleep(ms) { return new Promise(r => setTimeout(r, ms)); }

    // =========================================================================
    // مرحله ۱: صفحه کالا — افزودن به سبد خرید
    // =========================================================================
    if (window.location.pathname.startsWith('/product/')) {
        playAlertSound(880);
        showBanner('مرحله ۱/۴: در حال افزودن خودکار به سبد خرید...');

        let clicked = false;
        function tryAddToCart() {
            if (clicked) return;
            const btn = findByText(['افزودن به سبد خرید', 'افزودن به سبد'])
                     || document.querySelector('[data-cro-id="pdp-add-to-cart"]');
            if (btn) {
                clicked = true;
                btn.click();
                playAlertSound(1320);
                showBanner('✅ به سبد اضافه شد! در حال انتقال به ثبت سفارش...');
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
    // مرحله ۲: صفحه سبد — ثبت سفارش
    // =========================================================================
    if (window.location.pathname === '/checkout/cart') {
        playAlertSound(980);
        showBanner('مرحله ۲/۴: در حال ثبت سفارش...');

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
    // مرحله ۳: صفحه ارسال — SEQUENCE: سریعترین ارسال → اولین روز (label) → اولین ساعت (label/radio) → تایید و ادامه
    // =========================================================================
    if (window.location.pathname.includes('/checkout/shipping')) {
        playAlertSound(1100);
        showBanner('مرحله ۳/۴: انتخاب خودکار سریعترین ارسال...');

        // State machine: 0=click fastest tab, 1=pick day, 2=pick time, 3=confirm
        let phase = 0;
        let phaseTs = Date.now();
        let continueClicked = false;

        function phaseDelay(ms) {
            // only advance phase after a minimum delay so React can re-render
            return Date.now() - phaseTs >= ms;
        }

        function runShipping() {
            if (continueClicked) return;

            // ── Phase 0: کلیک روی تب «سریعترین ارسال» ───────────────────────
            if (phase === 0) {
                const fastTab = findByText(
                    ['سریع‌ترین ارسال', 'سریعترین ارسال', 'سریع ترین ارسال'],
                    'button, div[role="button"], label, div[class*="tab"], div[class*="chip"]'
                );
                if (fastTab) {
                    fastTab.click();
                    phase = 1;
                    phaseTs = Date.now();
                    showBanner('مرحله ۳/۴: سریعترین ارسال انتخاب شد — در حال انتخاب نزدیک‌ترین تاریخ...');
                }
                // also try to continue even if tab not found (maybe already on fastest)
                else if (phaseDelay(1200)) {
                    phase = 1;
                    phaseTs = Date.now();
                }
                return;
            }

            // ── Phase 1: انتخاب اولین روز فعال — target <label> elements ─────
            if (phase === 1 && phaseDelay(400)) {
                // Target label elements that contain day names (فردا / دوشنبه / ...)
                const dayLabels = Array.from(document.querySelectorAll('label')).filter(el => {
                    const txt = (el.textContent || '').trim();
                    return /فردا|دوشنبه|سه.?شنبه|چهارشنبه|پنجشنبه|جمعه|شنبه|یکشنبه/.test(txt)
                        && txt.length < 60
                        && !el.classList.contains('disabled')
                        && !el.closest('[class*="disabled"]')
                        && !el.closest('[aria-disabled="true"]');
                });

                // Fallback: div/button with day name and reasonable text length
                const dayFallback = dayLabels.length === 0
                    ? Array.from(document.querySelectorAll('div, button')).filter(el => {
                        const txt = (el.textContent || '').trim();
                        return /فردا|دوشنبه|سه.?شنبه|چهارشنبه|پنجشنبه|جمعه|شنبه|یکشنبه/.test(txt)
                            && txt.length < 50
                            && el.children.length <= 4
                            && !el.classList.contains('disabled')
                            && !el.closest('[class*="disabled"]');
                    })
                    : [];

                const dayEls = dayLabels.length > 0 ? dayLabels : dayFallback;
                if (dayEls.length > 0) {
                    dayEls[0].click();
                    phase = 2;
                    phaseTs = Date.now();
                    showBanner('مرحله ۳/۴: تاریخ ارسال انتخاب شد — در حال انتخاب اولین ساعت...');
                }
                return;
            }

            // ── Phase 2: انتخاب اولین ساعت — label > radio > div ─────────────
            if (phase === 2 && phaseDelay(400)) {
                // 1. Try label elements matching time ranges like «۱۱ الی ۱۴»
                const timeLabels = Array.from(document.querySelectorAll('label')).filter(el => {
                    const txt = (el.textContent || '').trim();
                    return /\d+\s*الی\s*\d+/.test(txt)
                        && txt.length < 60
                        && !el.classList.contains('disabled')
                        && !el.closest('[class*="disabled"]')
                        && !el.closest('[aria-disabled="true"]');
                });

                if (timeLabels.length > 0) {
                    timeLabels[0].click();
                    phase = 3;
                    phaseTs = Date.now();
                    showBanner('مرحله ۳/۴: ساعت ارسال انتخاب شد — در حال تایید...');
                    return;
                }

                // 2. Fallback: enabled radio buttons
                const radios = Array.from(document.querySelectorAll('input[type="radio"]:not(:disabled)'));
                if (radios.length > 0) {
                    radios[0].click();
                    if (!radios[0].checked) {
                        radios[0].checked = true;
                        radios[0].dispatchEvent(new Event('change', { bubbles: true }));
                    }
                    phase = 3;
                    phaseTs = Date.now();
                    showBanner('مرحله ۳/۴: ساعت ارسال انتخاب شد — در حال تایید...');
                    return;
                }

                // 3. Fallback: div/span with time range text
                const timeEl = Array.from(document.querySelectorAll('div, span')).find(el => {
                    const txt = (el.textContent || '').trim();
                    return /\d+\s*الی\s*\d+/.test(txt)
                        && txt.length < 40
                        && el.children.length <= 3
                        && !el.closest('[class*="disabled"]');
                });
                if (timeEl) {
                    timeEl.click();
                    phase = 3;
                    phaseTs = Date.now();
                    showBanner('مرحله ۳/۴: ساعت ارسال انتخاب شد — در حال تایید...');
                }
                return;
            }

            // ── Phase 3: تایید و ادامه ────────────────────────────────────────
            if (phase === 3 && phaseDelay(500)) {
                const confirmBtn = findByText(['تایید و ادامه', 'ادامه و تایید', 'ادامه جهت پرداخت']);
                if (confirmBtn && !confirmBtn.disabled && !confirmBtn.classList.contains('disabled')) {
                    continueClicked = true;
                    confirmBtn.click();
                    playAlertSound(1500);
                    showBanner('✅ زمان ارسال تایید شد! انتقال به صفحه پرداخت...');
                    setTimeout(() => {
                        if (!window.location.pathname.includes('/payment')) {
                            window.location.href = 'https://snappshop.ir/checkout/payment?auto_pay=1';
                        }
                    }, 900);
                }
            }
        }

        const iv = setInterval(runShipping, 60);
        new MutationObserver(runShipping).observe(document.documentElement, { childList: true, subtree: true });
        setTimeout(() => clearInterval(iv), 20000);
    }

    // =========================================================================
    // مرحله ۴: صفحه پرداخت — کلیک «پرداخت از طریق درگاه بانکی»
    // =========================================================================
    if (window.location.pathname.includes('/checkout/payment')) {
        playAlertSound(1200);
        showBanner('مرحله ۴/۴: در حال کلیک روی پرداخت از طریق درگاه بانکی...');

        let done = false;
        function tryPay() {
            if (done) return;
            const btn = findByText([
                'پرداخت از طریق درگاه بانکی',
                'پرداخت از طریق درگاه',
                'پرداخت اینترنتی',
                'پرداخت آنلاین'
            ]);
            if (btn && !btn.disabled && !btn.classList.contains('disabled')) {
                done = true;
                sessionStorage.removeItem('snapp_quick_buy');
                sessionStorage.removeItem('snapp_s26_sniper');
                btn.click();
                playAlertSound(1760);
                showBanner('🚀 انتقال به درگاه بانکی! لطفاً رمز پویا و اطلاعات کارت را وارد کنید.',
                    'linear-gradient(90deg, #16a34a, #15803d)');
            }
        }

        const iv = setInterval(tryPay, 50);
        new MutationObserver(tryPay).observe(document.documentElement, { childList: true, subtree: true });
        setTimeout(() => clearInterval(iv), 15000);
    }

})();
