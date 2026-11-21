import urllib.request, re, urllib.parse
H = {'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', 'Referer': 'https://snappshop.ir/'}
def get(u):
    return urllib.request.urlopen(urllib.request.Request(u, headers=H), timeout=15).read().decode('utf-8', 'replace')

pages = ['https://snappshop.ir/', 'https://snappshop.ir/search?query=' + urllib.parse.quote('هارد'), 'https://snappshop.ir/discounts']
chunks = set()
for p in pages:
    try:
        h = get(p)
    except Exception as e:
        print('page fail', p, e); continue
    found = re.findall(r'/_next/static/[^"\'\s\\]+\.js', h)
    print(p, len(h), 'chunks', len(found))
    chunks |= set(found)
print('total chunks', len(chunks))
hits = 0
for c in sorted(chunks):
    try:
        js = get('https://snappshop.ir' + c)
    except Exception:
        continue
    for kw in ['per_page', 'total_pages', 'search/v1', 'has_discount', 'is_available']:
        for m in re.finditer(kw, js):
            s = max(0, m.start() - 150); e = min(len(js), m.end() + 200)
            print('---', c.split('/')[-1], kw)
            print(js[s:e].replace('\n', ' '))
            hits += 1
            break
        if hits > 12: break
    if hits > 12: break
print('hits', hits)
