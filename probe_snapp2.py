import urllib.request, urllib.error, json, time

H = {'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
     'Content-Type': 'application/json', 'Accept': 'application/json',
     'Origin': 'https://snappshop.ir', 'Referer': 'https://snappshop.ir/'}

def call(extra, url_extra=''):
    b = {'sort': '5DzQq4', 'is_available': True, 'has_discount': True, 'limit': 100}
    b.update(extra)
    r = urllib.request.Request('https://apix.snappshop.ir/search/v1?lat=35.77331&lng=51.418591' + url_extra,
                               data=json.dumps(b).encode(), headers=H)
    try:
        with urllib.request.urlopen(r, timeout=15) as f:
            d = json.loads(f.read())
    except urllib.error.HTTPError as e:
        return 'HTTP %d' % e.code, None
    items = []
    meta = {}
    for s in d['data']['structure']:
        items += s.get('items', [])
        meta = {k: v for k, v in s.items() if k not in ('items', 'filter')}
    return [i['id'] for i in items], meta

base, meta = call({'page': 1})
print('page1', len(base), 'meta:', meta)
variants = {
  'page=2': ({'page': 2}, ''),
  'skip=100': ({'page': 1, 'skip': 100}, ''),
  'offset=100': ({'page': 1, 'offset': 100}, ''),
  'url page=2': ({}, '&page=2'),
  'page=2 limit=20': ({'page': 2, 'limit': 20}, ''),
  'page=0': ({'page': 0}, ''),
}
for name, (extra, ue) in variants.items():
    time.sleep(1.5)
    ids, m = call(extra, ue)
    if isinstance(ids, str):
        print(name, ids); continue
    print(name, 'n=%d overlap_with_p1=%d' % (len(ids), len(set(ids) & set(base))))
