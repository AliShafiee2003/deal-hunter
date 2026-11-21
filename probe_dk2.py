import urllib.request, urllib.parse, http.cookiejar, json

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

d = get_json('https://api.digikala.com/discovery/api/v2/search?q=' + urllib.parse.quote('لباس') + '&sort=26&has_selling_stock=1&page=1')
w0 = d['data']['widgets'][0]['data']
print("=== FILTERS ===")
for k, v in w0.get('filters', {}).items():
    print(f"Filter key: {k} -> type={v.get('type')}, title={v.get('title')}")

print("\n=== SORT OPTIONS ===")
print(w0.get('sort_options'))

# Test other discovery endpoints!
endpoints = [
    'https://api.digikala.com/discovery/api/v2/search?has_selling_stock=1&has_ready_to_shipment=1&sort=26&page=1',
    'https://api.digikala.com/discovery/api/v2/search?q=&has_selling_stock=1&sort=26&page=1',
    'https://api.digikala.com/discovery/api/v2/categories/mobile-phone/search?sort=26&has_selling_stock=1&page=1',
    'https://api.digikala.com/discovery/api/v2/category/mobile-phone/?sort=26&has_selling_stock=1&page=1',
    'https://api.digikala.com/discovery/api/v2/incredible-offers/',
    'https://api.digikala.com/v1/incredible-offers/',
    'https://api.digikala.com/v2/incredible-offers/',
    'https://api.digikala.com/v1/categories/mobile-phone/search/?sort=26&has_selling_stock=1&page=1',
]

print("\n=== ENDPOINT PROBES ===")
for ep in endpoints:
    try:
        res = get_json(ep)
        st = res.get('status')
        dk = list(res.get('data', {}).keys()) if isinstance(res.get('data'), dict) else type(res.get('data'))
        # count products and max/min discount if widgets exist
        cnt, maxd, mind = 0, 0, 0
        if isinstance(res.get('data'), dict) and 'widgets' in res['data']:
            discs = []
            for w in res['data']['widgets']:
                if w.get('type') == 'vertical_product_listing':
                    for it in w.get('data', {}).get('widgets', []):
                        if it.get('type') == 'product':
                            cnt += 1
                            discs.append(it['data'].get('default_variant', {}).get('price', {}).get('discount_percent', 0))
            if discs: maxd, mind = max(discs), min(discs)
        print(f"OK [{st}] {ep} -> data_keys={dk} products={cnt} disc={maxd}..{mind}")
    except Exception as e:
        print(f"ERR {ep} -> {e}")
