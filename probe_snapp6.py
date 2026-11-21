import urllib.request, re
H = {'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', 'Referer': 'https://snappshop.ir/'}
def get(u):
    return urllib.request.urlopen(urllib.request.Request(u, headers=H), timeout=15).read().decode('utf-8', 'replace')
h = get('https://snappshop.ir/')
chunks = sorted(set(re.findall(r'/_next/static/[^"\'\s\\]+\.js', h)))
for c in chunks:
    try: js = get('https://snappshop.ir' + c)
    except Exception: continue
    for pat in [r'skip\s*:', r'"skip"', r'page_number', r'\.skip\b', r'offset\s*:']:
        for m in list(re.finditer(pat, js))[:2]:
            s = max(0, m.start() - 200); e = min(len(js), m.end() + 250)
            print('---', c.split('/')[-1], pat)
            print(js[s:e].replace('\n', ' '))
