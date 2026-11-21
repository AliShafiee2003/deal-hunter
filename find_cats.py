import urllib.request, re, time, json

headers = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
    'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
    'Accept-Language': 'fa-IR,fa;q=0.9,en;q=0.8',
    'Referer': 'https://snappshop.ir/'
}

print("Fetching SnappShop homepage...")
req = urllib.request.Request('https://snappshop.ir/', headers=headers)
with urllib.request.urlopen(req, timeout=12) as r:
    html = r.read().decode('utf-8', errors='replace')

print(f"Page size: {len(html)} chars")

# Look for category codes in /category/XXXX pattern
hrefs = re.findall(r'/category/([A-Za-z0-9]{4,12})', html)
unique = list(dict.fromkeys(hrefs))
print(f"\nCategory codes from hrefs ({len(unique)}):")
for c in unique[:60]:
    print(f"  {c}")

# Look for Next.js data
next_data_match = re.search(r'<script id="__NEXT_DATA__"[^>]*>(.*?)</script>', html, re.DOTALL)
if next_data_match:
    try:
        data = json.loads(next_data_match.group(1))
        txt = json.dumps(data, ensure_ascii=False)
        # Find category entries - look for "id" near "name" in Persian
        # Extract category objects
        cat_objs = re.findall(r'"id"\s*:\s*"([A-Za-z0-9]{4,12})".*?"name"\s*:\s*"([^"]{2,30})"', txt[:50000])
        print(f"\nCategory id+name pairs ({len(cat_objs)}):")
        for cid, cname in cat_objs[:40]:
            print(f"  {cid}: {cname}")
    except Exception as e:
        print(f"NEXT_DATA parse error: {e}")
else:
    print("No __NEXT_DATA__ found")

# Also search for category name mappings from aria-label or title attributes
titles = re.findall(r'href="/category/([A-Za-z0-9]{4,12})"[^>]*(?:title|aria-label)="([^"]+)"', html)
print(f"\nCategory href+title ({len(titles)}):")
for code, name in titles[:30]:
    print(f"  {code}: {name}")
