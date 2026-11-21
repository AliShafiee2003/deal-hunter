import urllib.request, json, time

headers = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
    'Content-Type': 'application/json',
    'Accept': 'application/json',
    'Origin': 'https://snappshop.ir',
    'Referer': 'https://snappshop.ir/'
}

LAT = "35.77331"
LNG = "51.418591"

# From the screenshot, main categories visible in SnappShop:
# موبایل, دیجیتال, زیبایی و سلامت, خانه و آشپزخانه, لوازم خانگی برقی, مد و پوشاک, طلا و نقره, فرهنگ و هنر

# Try to search for each category name and get the category_id from results
test_queries = [
    ("زیبایی و سلامت", "beauty_health"),
    ("خانه و آشپزخانه", "home_kitchen"),
    ("لوازم خانگی برقی", "home_appliance"),
    ("مد و پوشاک", "fashion"),
    ("طلا و نقره", "gold_silver"),
    ("فرهنگ و هنر", "culture_art"),
    ("اسباب بازی", "toy"),
    ("دوربین", "camera"),
    ("تبلت", "tablet"),
    ("ورزشی", "sports"),
    ("کتاب", "book"),
    ("خودرو", "auto"),
    ("ابزار", "tools"),
]

found = {}
base_url = f"https://apix.snappshop.ir/search/v1?lat={LAT}&lng={LNG}"

for query, key in test_queries:
    time.sleep(1.2)  # avoid rate limit
    body = json.dumps({
        "query": query,
        "sort": "5DzQq4",
        "is_available": True,
        "has_discount": True,
        "page": 1,
        "limit": 3
    }).encode()
    try:
        req = urllib.request.Request(base_url, data=body, headers=headers)
        with urllib.request.urlopen(req, timeout=10) as r:
            data = json.loads(r.read())
        sections = data.get('data', {}).get('structure', [])
        items = []
        for s in sections:
            items.extend(s.get('items', []))
        if items:
            item = items[0]
            href = item.get('href', '')
            title = item.get('title', '')
            # href pattern: /category/CATCODE/product/...  or  /product/snp-...
            import re
            cat_match = re.search(r'/category/([A-Za-z0-9]+)', href)
            cat_code = cat_match.group(1) if cat_match else "N/A (no cat in URL)"
            print(f"[{key}] query='{query}' => first item: '{title[:40]}', href={href[:60]}, cat={cat_code}")
            found[key] = cat_code
        else:
            print(f"[{key}] query='{query}' => NO RESULTS")
    except Exception as e:
        print(f"[{key}] query='{query}' => ERROR: {e}")

print("\n\nSummary of found category codes:")
for k, v in found.items():
    print(f"  {k}: {v}")
