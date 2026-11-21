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

# 1. Inspect categories filter structure
d = get_json('https://api.digikala.com/discovery/api/v2/search?has_selling_stock=1&has_ready_to_shipment=1&sort=26&page=1')
w0 = d['data']['widgets'][0]['data']
print("Pager (no q, has_ready_to_shipment=1):", w0.get('pager'))
print("Categories filter:", json.dumps(w0.get('filters', {}).get('categories'), ensure_ascii=False)[:1500])

# 2. Check the 5 user products directly via product API or search to see their exact categories & discounts
dkps = [6262614, 23263447, 19184461, 22679760, 23055189]
print("\n=== USER 5 PRODUCTS ===")
for pid in dkps:
    for ep in [f'https://api.digikala.com/v2/product/{pid}/', f'https://api.digikala.com/v1/product/{pid}/']:
        try:
            pd = get_json(ep)
            p = pd.get('data', {}).get('product', {})
            if p:
                cat = p.get('category', {})
                dl = p.get('data_layer', {})
                pr = p.get('default_variant', {}).get('price', {}) if isinstance(p.get('default_variant'), dict) else {}
                print(f"DKP-{pid}: title='{p.get('title_fa')}' | cat_id={cat.get('id')} cat_code={cat.get('code')} cat_title={cat.get('title_fa')} | dl_cat={dl.get('item_category2')}/{dl.get('item_category3')} | disc={pr.get('discount_percent')}%")
                break
        except Exception as e:
            print(f"DKP-{pid} ({ep}): {e}")

# 3. Test category filter parameters in discovery/api/v2/search
print("\n=== CATEGORY FILTER PARAM TESTS ===")
test_urls = [
    'https://api.digikala.com/discovery/api/v2/search?categories[]=11&sort=26&has_selling_stock=1&page=1',
    'https://api.digikala.com/discovery/api/v2/search?categories[0]=11&sort=26&has_selling_stock=1&page=1',
    'https://api.digikala.com/discovery/api/v2/search?category_id=11&sort=26&has_selling_stock=1&page=1',
    'https://api.digikala.com/discovery/api/v2/search?has_selling_stock=1&sort=26&page=1',
    'https://api.digikala.com/discovery/api/v2/search?has_selling_stock=1&sort=26&page=2',
    'https://api.digikala.com/discovery/api/v2/search?has_selling_stock=1&sort=26&page=5',
    'https://api.digikala.com/discovery/api/v2/search?has_selling_stock=1&sort=26&page=10',
    'https://api.digikala.com/discovery/api/v2/search?has_selling_stock=1&sort=26&page=20',
]
for u in test_urls:
    try:
        r = get_json(u)
        w = r['data']['widgets'][0]['data']
        items = [it['data'] for it in w.get('widgets', []) if it.get('type') == 'product']
        discs = [it.get('default_variant', {}).get('price', {}).get('discount_percent', 0) for it in items]
        first_title = items[0].get('title_fa', '')[:35] if items else ''
        print(f"URL: {u.split('?')[1]} -> total={w.get('pager',{}).get('total_items')} disc={max(discs,default=0)}..{min(discs,default=0)} first='{first_title}'")
    except Exception as e:
        print(f"URL ERR {u}: {e}")
