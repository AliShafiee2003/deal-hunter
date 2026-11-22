// ==UserScript==
// @name         Fast Buy 1-Click for SnappShop (Complete Auto-Checkout to Bank Gateway)
// @namespace    https://snappshop.ir/
// @version      5.1
// @description  اسنایپر متمرکز بر قیمت و تخفیف: رصد خودکار کالای منحصراً با تخفیف ۹۰٪ (بج بنفش ٪۹۰) و قیمت ۴ میلیون و ۴۰۰ هزار تومان در صفحه ویژه Doe3z0، اسکرول عمیق لیزی‌لود و رفرش هر ۱۰ ثانیه، افزودن به سبد، ثبت سفارش، انتخاب سریعترین ارسال و اتصال مستقیم به درگاه بانکی
// @match        https://snappshop.ir/*
// @match        https://www.snappshop.ir/*
// @run-at       document-start
// @grant        none
// ==/UserScript==

(function() {
    'use strict';

    const urlParams = new URLSearchParams(window.location.search);
    const isQuickBuy = urlParams.get('quick_buy') === '1' || sessionStorage.getItem('snapp_quick_buy') === '1';
    const isSniper = urlParams.get('sniper') === 'deal440' || urlParams.get('sniper') === 's26' || sessionStorage.getItem('snapp_sniper') === '1';
    const isDoe3z0 = window.location.pathname.includes('/product-list/Doe3z0');
    const isProductListPage = window.location.pathname.startsWith('/product-list/');

    if (isSniper)   sessionStorage.setItem('snapp_sniper', '1');
    if (isQuickBuy) sessionStorage.setItem('snapp_quick_buy', '1');

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
    const defaultTheme = 'linear-gradient(90deg, #dc2626, #b91c1c)';

    function showBanner(message, bgColor) {
        bgColor = bgColor || defaultTheme;
        let banner = document.getElementById('fast-buy-banner');
        if (!banner) {
            banner = document.createElement('div');
            banner.id = 'fast-buy-banner';
            banner.style.cssText = 'position:fixed;top:0;left:0;right:0;color:#fff;text-align:center;padding:14px 10px;font-size:15px;font-weight:bold;z-index:99999999;box-shadow:0 4px 25px rgba(0,0,0,0.7);direction:rtl;font-family:system-ui,tahoma,sans-serif;';
            const attach = () => document.body && document.body.appendChild(banner);
            if (document.body) attach(); else document.addEventListener('DOMContentLoaded', attach);
        }
        banner.style.background = bgColor;
        banner.innerHTML = '🎯 [اسنایپر شکار آفر ویژه ۹۰٪] ' + message;
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
    // استخراج دقیق درصد تخفیف از برچسب کارت محصول
    // بررسی بج بنفش/قرمز تخفیف (مانند %70 ، %64 ، %90 ، ٪۹۰)
    // -------------------------------------------------------------------------
    function getCardDiscountPercent(card) {
        // ۱. بررسی المان‌های برگ (بدون فرزند) داخل کارت محصول
        const allEls = Array.from(card.querySelectorAll('*'));
        for (const el of allEls) {
            if (el.children.length === 0) {
                const raw = (el.textContent || '').trim();
                if (raw.includes('%') || raw.includes('٪')) {
                    const norm = normalizeDigits(raw).replace(/\s+/g, '');
                    // تطابق الگوهایی مانند %70 یا 70% یا ٪64 یا %90
                    const m = norm.match(/^[%٪]?(\d{1,2})[%٪]?$/);
                    if (m) {
                        const val = parseInt(m[1], 10);
                        if (val > 0 && val <= 100) {
                            return val;
                        }
                    }
                }
            }
        }

        // ۲. پشتیبانی از کلاس‌های مربوط به بج تخفیف
        const badgeEl = card.querySelector('[class*="badge"], [class*="discount"], [class*="percent"], [class*="tag"]');
        if (badgeEl) {
            const raw = normalizeDigits(badgeEl.textContent || '').replace(/\s+/g, '');
            const m = raw.match(/(\d{1,2})[%٪]|[%٪](\d{1,2})/);
            if (m) {
                const val = parseInt(m[1] || m[2], 10);
                if (val > 0 && val <= 100) return val;
            }
        }

        // ۳. بررسی کلی روی متن کارت به عنوان آخرین راهکار
        const fullNorm = normalizeDigits(card.textContent || '').replace(/\s+/g, '');
        const generalMatch = fullNorm.match(/[%٪](\d{1,2})|(\d{1,2})[%٪]/);
        if (generalMatch) {
            const val = parseInt(generalMatch[1] || generalMatch[2], 10);
            if (val > 0 && val <= 100) return val;
        }

        return 0;
    }

    // -------------------------------------------------------------------------
    // استخراج دقیق قیمت نهایی فروش از کارت محصول
    // بررسی ۴,۴۰۰,۰۰۰ تومان (یا معادل ریال ۴۴,۰۰۰,۰۰۰)
    // -------------------------------------------------------------------------
    function getCardPrice(card) {
        const raw = normalizeDigits(card.textContent || '');
        const clean = raw.replace(/[,،\s]/g, '');

        // ۱. تطابق مستقیم ۴,۴۰۰,۰۰۰ تومان یا ۴۴,۰۰۰,۰۰۰ ریال
        if (clean.includes('4400000') || clean.includes('44000000')) {
            return 4400000;
        }

        // ۲. بررسی تمام اعداد ۶ رقمی به بالا
        const numbers = clean.match(/\d{6,10}/g);
        if (numbers) {
            for (const numStr of numbers) {
                const n = parseInt(numStr, 10);
                // بازه تومان: ۴,۳۵۰,۰۰۰ تا ۴,۴۵۰,۰۰۰
                if (n >= 4350000 && n <= 4450000) {
                    return 4400000;
                }
                // بازه ریال: ۴۳,۵۰۰,۰۰۰ تا ۴۴,۵۰۰,۰۰۰
                if (n >= 43500000 && n <= 44500000) {
                    return 4400000;
                }
            }
        }
        return 0;
    }

    // -------------------------------------------------------------------------
    // Target Matcher Logic: زوم انحصاری روی تخفیف ۹۰٪ + قیمت ۴,۴۰۰,۰۰۰ تومان
    // -------------------------------------------------------------------------
    function isMatchingTargetDeal(card, title) {
        const disc = getCardDiscountPercent(card);
        const price = getCardPrice(card);

        // لاگ کمکی در کنسول
        if (disc > 0 || price > 0) {
            console.log(`[Sniper Monitor] "${(title || '').slice(0, 30)}..." => تخفیف: ${disc}% | قیمت: ${price}`);
        }

        // شرط قطعی و حیاتی:
        // ۱. درصد تخفیف حتماً باید ۹۰٪ باشد (۸۹ تا ۹۲ درصد)
        // ۲. قیمت حتماً باید ۴ میلیون و ۴۰۰ هزار تومان باشد
        // به هیچ وجه کالاهای با تخفیف ۶۴٪، ۷۰٪ و... قبول نخواهند شد!
        const is90Discount = (disc >= 89 && disc <= 92);
        const is44Price = (price === 4400000);

        if (is90Discount && is44Price) {
            console.log(`🎯 [SNIPER MATCH!] کالای هدف ۹۰٪ پیدا شد: تخفیف=${disc}%، قیمت=${price}`);
            return true;
        }

        return false;
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
    // مرحله ۰: پایش ویژه لیست محصولات (product-list/*)
    // اسکرول عمیق ۲ برابری برای Lazy-Load + رفرش خودکار هر ۱۰ ثانیه
    // =========================================================================
    if (isProductListPage) {
        let dealFound = false;
        const pageLoadTs = Date.now();
        let scrollStep = 0;

        function scanProductCards() {
            if (dealFound) return;
            const cards = Array.from(document.querySelectorAll('a[href*="/product/snp-"]'));

            for (const card of cards) {
                const titleEl = card.querySelector('h2, h3, h4, p, [class*="title"]') || card;
                const title = titleEl.textContent || '';

                if (isMatchingTargetDeal(card, title)) {
                    dealFound = true;
                    playAlertSound(1500);
                    showBanner('🚨 کالای ۹۰٪ تخفیف (۴,۴۰۰,۰۰۰ تومان) پیدا شد! انتقال فوری به خرید...', 'linear-gradient(90deg, #16a34a, #15803d)');

                    let targetHref = card.getAttribute('href') || '';
                    if (!targetHref.startsWith('http')) {
                        targetHref = 'https://snappshop.ir' + targetHref;
                    }
                    const sep = targetHref.includes('?') ? '&' : '?';
                    window.location.href = targetHref + sep + 'quick_buy=1&sniper=deal440';
                    return;
                }
            }
        }

        const scanIv = setInterval(scanProductCards, 40);
        new MutationObserver(scanProductCards).observe(document.documentElement, { childList: true, subtree: true });
        window.addEventListener('scroll', scanProductCards, { passive: true });

        // اتوماسیون مداوم اسکرول و رفرش هر ۱۰ ثانیه
        function automateListPage() {
            if (dealFound) return;

            const pageLabel = isDoe3z0 ? 'صفحه ویژه Doe3z0' : 'لیست محصولات';
            showBanner(`🔥 اسنایپر ${pageLabel} فعال است: رصد بج تخفیف ۹۰٪ + قیمت ۴.۴ میلیون (اسکرول عمیق و رفرش هر ۱۰ ثانیه)...`);

            // ۱. اسکرول عمیق و پرشتاب به پایین (۲ برابر مقدار استاندارد) برای باز کردن فوری لایه‌های لیزی‌لود
            scrollStep++;
            const scrollPositions = [2000, 4500, 8000, 12000, document.body.scrollHeight];
            const targetY = scrollPositions[scrollStep % scrollPositions.length] || document.body.scrollHeight;
            window.scrollTo({ top: targetY, behavior: 'smooth' });

            // ۲. رفرش صفحه هر ۱۰ ثانیه برای دریافت کالاهای تازه اضافه‌شده
            const elapsed = Date.now() - pageLoadTs;
            if (elapsed >= 10000) {
                scanProductCards();
                if (!dealFound) {
                    window.location.reload();
                }
            }
        }

        setInterval(automateListPage, 700);
    }

    // =========================================================================
    // مرحله ۱: صفحه کالا — راستی‌آزمایی ایمنی تخفیف + افزودن به سبد خرید
    // =========================================================================
    if (window.location.pathname.startsWith('/product/') && (isQuickBuy || isSniper)) {
        playAlertSound(880);
        showBanner('مرحله ۱/۴: بررسی اطلاعات کالا و افزودن خودکار به سبد خرید...');

        // استخراج درصد تخفیف در صفحه کالا جهت جلوگیری از خطای خرید کالاهای نامربوط
        function getPDPDiscountPercent() {
            const allEls = Array.from(document.querySelectorAll('*'));
            for (const el of allEls) {
                if (el.children.length === 0) {
                    const raw = (el.textContent || '').trim();
                    if (raw.includes('%') || raw.includes('٪')) {
                        const norm = normalizeDigits(raw).replace(/\s+/g, '');
                        const m = norm.match(/^[%٪]?(\d{1,2})[%٪]?$/);
                        if (m) {
                            const val = parseInt(m[1], 10);
                            if (val > 0 && val <= 100) return val;
                        }
                    }
                }
            }
            return 0;
        }

        let clicked = false;
        function tryAddToCart() {
            if (clicked) return;

            // راستی‌آزمایی ایمنی: اگر تخفیف کالا کمتر از ۸۸٪ باشد (مثلاً ۶۴٪ کتاب یا ۷۰٪)، خرید فوراً لغو می‌شود!
            if (isSniper) {
                const pdpDisc = getPDPDiscountPercent();
                if (pdpDisc > 0 && (pdpDisc < 88 || pdpDisc > 95)) {
                    clicked = true;
                    sessionStorage.removeItem('snapp_sniper');
                    sessionStorage.removeItem('snapp_quick_buy');
                    showBanner(`⚠️ توقف اضطراری: تخفیف این کالا ${pdpDisc}٪ است (تخفیف ۹۰٪ نیست)! خرید لغو شد.`, 'linear-gradient(90deg, #b91c1c, #7f1d1d)');
                    return;
                }
            }

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
    if (window.location.pathname === '/checkout/cart' && (isQuickBuy || isSniper)) {
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
    if (window.location.pathname.includes('/checkout/shipping') && (isQuickBuy || isSniper)) {
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
    if (window.location.pathname.includes('/checkout/payment') && (isQuickBuy || isSniper)) {
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
                sessionStorage.removeItem('snapp_sniper');
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
