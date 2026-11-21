import urllib.request, urllib.parse, http.cookiejar, json, time

cj = http.cookiejar.CookieJar()
opener = urllib.request.build_opener(urllib.request.HTTPCookieProcessor(cj))
H = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
    'Accept': 'application/json, text/plain, */*',
    'Accept-Language': 'fa-IR,fa;q=0.9,en;q=0.8',
    'Origin': 'https://www.digikala.com',
    'Referer': 'https://www.digikala.com/',
    'x-web-client': 'desktop',
    'x-web-client-id': 'web',
}

def get_json(url):
    req = urllib.request.Request(url, headers=H)
    with opener.open(req, timeout=15) as r:
        return json.loads(r.read())

# 1. Inspect full response of discovery search for pagination & filters
d = get_json('https://api.digikala.com/discovery/api/v2/search?q=' + urllib.parse.quote('گوشی موبایل') + '&sort=26&has_selling_stock=1&page=1')
print("Top keys:", list(d.keys()))
print("Data keys:", list(d.get('data', {}).keys()))
for i, w in enumerate(d.get('data', {}).get('widgets', [])):
    wtype = w.get('type')
    wdata = w.get('data', {})
    print(f"Widget {i}: type={wtype}, keys={list(wdata.keys()) if isinstance(wdata, dict) else type(wdata)}")
    if wtype == 'vertical_product_listing':
        print("  vpl pager/meta:", {k: v for k, v in wdata.items() if k != 'widgets'})
        print("  vpl count:", len(wdata.get('widgets', [])))

# 2. Test pages 1 to 15 of a single query to see if pagination really advances
seen = set()
for p in [1, 2, 5, 10, 15, 20]:
    d2 = get_json(f'https://api.digikala.com/discovery/api/v2/search?q={urllib.parse.quote("گوشی")}&sort=26&has_selling_stock=1&page={p}')
    ids = []
    discs = []
    for w in d2.get('data', {}).get('widgets', []):
        if w.get('type') == 'vertical_product_listing':
            for it in w.get('data', {}).get('widgets', []):
                if it.get('type') == 'product':
                    pid = it['data'].get('id')
                    pr = it['data'].get('default_variant', {}).get('price', {})
                    ids.append(pid)
                    discs.append(pr.get('discount_percent', 0))
    new = len(set(ids) - seen)
    seen |= set(ids)
    print(f"Page {p}: count={len(ids)}, new={new}, max_disc={max(discs, default=0)}, min_disc={min(discs, default=0)}")
