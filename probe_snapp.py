import urllib.request, urllib.error, json, time

H = {'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
     'Content-Type': 'application/json', 'Accept': 'application/json',
     'Origin': 'https://snappshop.ir', 'Referer': 'https://snappshop.ir/'}
URL = 'https://apix.snappshop.ir/search/v1?lat=35.77331&lng=51.418591'

def page(p, query='', cat=''):
    b = {'sort': '5DzQq4', 'is_available': True, 'has_discount': True, 'page': p, 'limit': 100}
    if query: b['query'] = query
    if cat: b['category'] = cat
    r = urllib.request.Request(URL, data=json.dumps(b).encode(), headers=H)
    with urllib.request.urlopen(r, timeout=15) as f:
        d = json.loads(f.read())
    items = []
    for s in d['data']['structure']:
        items += s.get('items', [])
    return items

def disc(it):
    p = it.get('price', {})
    return p.get('discount') or p.get('discount_percent') or 0

seen = set()
print("== unfiltered, has_discount, sorted by discount ==")
for p in range(1, 31):
    try:
        items = page(p)
    except urllib.error.HTTPError as e:
        print(f"page {p}: HTTP {e.code}")
        if e.code == 429:
            break
        continue
    except Exception as e:
        print(f"page {p}: {e}")
        break
    ids = {i['id'] for i in items}
    new = len(ids - seen)
    seen |= ids
    ds = [disc(i) for i in items]
    print(f"page {p}: n={len(items)} new={new} disc max={max(ds, default=0)} min={min(ds, default=0)} >=70: {sum(d>=70 for d in ds)}")
    if not items:
        break
    time.sleep(1.5)
print("unique seen:", len(seen))
