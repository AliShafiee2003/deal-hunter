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

# 1. Inspect data_layer of products returned by unfiltered sort=26
d = get_json('https://api.digikala.com/discovery/api/v2/search?has_selling_stock=1&sort=26&page=1')
w0 = d['data']['widgets'][0]['data']
opts = w0.get('filters', {}).get('categories', {}).get('options', [])
print(f"Total category options in unfiltered search: {len(opts)}")

print("\nSample products from unfiltered sort=26 page 1:")
for it in w0.get('widgets', [])[:10]:
    if it.get('type') == 'product':
        p = it['data']
        dl = p.get('data_layer', {})
        pr = p.get('default_variant', {}).get('price', {})
        print(f"  [{pr.get('discount_percent')}%] {p.get('title_fa')[:35]} -> cat={dl.get('category')} | c2={dl.get('item_category2')} | c3={dl.get('item_category3')}")

# 2. Check query_attributes.category_ids for our main queries to discover main numeric category IDs!
queries = [
    'هارد اکسترنال', 'اس اس دی', 'گوشی موبایل', 'تبلت', 'لپ تاپ', 'کنسول بازی',
    'هدفون', 'اسپیکر', 'دوربین عکاسی', 'ساعت هوشمند', 'پاوربانک',
    'لوازم خانگی برقی', 'خانه و آشپزخانه', 'یخچال', 'جاروبرقی', 'ماشین لباسشویی', 'قهوه ساز', 'تمپر قهوه',
    'آرایشی بهداشتی', 'مکمل', 'عینک طبی', 'کپسول مکمل',
    'پوشاک', 'کفش', 'طلا',
    'ابزار', 'لوازم خودرو', 'سوپرمارکت', 'کتاب', 'ورزش و سفر', 'اسباب بازی'
]

print("\n=== NUMERIC CATEGORY IDs FROM QUERIES ===")
for q in queries:
    try:
        r = get_json('https://api.digikala.com/discovery/api/v2/search?q=' + urllib.parse.quote(q) + '&sort=26&has_selling_stock=1&page=1')
        w = r['data']['widgets'][0]['data']
        cids = w.get('query_attributes', {}).get('category_ids', [])
        cat_opts = w.get('filters', {}).get('categories', {}).get('options', [])[:3]
        cat_str = ", ".join([f"{c.get('id')}:{c.get('title_fa')}" for c in cat_opts])
        print(f"  q='{q}' -> query_attr_cids={cids} | top_filter_cats=[{cat_str}]")
    except Exception as e:
        print(f"  q='{q}' -> ERR: {e}")
