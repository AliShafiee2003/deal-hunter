"""
Extract ALL SnappShop category codes using the category_tree filter
returned by the search API when querying different category names.
"""
import urllib.request, json, time, re

headers = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    'Content-Type': 'application/json',
    'Accept': 'application/json',
    'Origin': 'https://snappshop.ir',
    'Referer': 'https://snappshop.ir/'
}

LAT = "35.77331"
LNG = "51.418591"
BASE = f"https://apix.snappshop.ir/search/v1?lat={LAT}&lng={LNG}"

# Main category terms to discover sub-category trees
# Each query returns its own category_tree filter
QUERIES = [
    "موبایل",
    "لپ تاپ",
    "هارد اکسترنال",
    "هدفون",
    "اسپیکر",
    "ساعت هوشمند",
    "پاوربانک",
    "دوربین",
    "تبلت",
    "لوازم خانگی",
    "یخچال",
    "جارو برقی",
    "لباسشویی",
    "آرایشی",
    "مکمل",
    "عینک",
    "لباس",
    "کفش",
    "طلا",
    "کتاب",
    "ابزار",
    "خودرو",
    "اسباب بازی",
    "ورزشی",
    "سوپرمارکت",
    "قهوه",
    "مبل",
    "کنسول",
]

all_cats = {}  # id -> name

for i, query in enumerate(QUERIES):
    if i > 0:
        time.sleep(2)
    
    body = json.dumps({
        "query": query,
        "sort": "5DzQq4",
        "is_available": True,
        "has_discount": True,
        "page": 1,
        "limit": 1
    }).encode()
    
    try:
        req = urllib.request.Request(BASE, data=body, headers=headers)
        with urllib.request.urlopen(req, timeout=10) as r:
            data = json.loads(r.read())
        
        txt = json.dumps(data, ensure_ascii=False)
        # Find category_tree filter items
        # Pattern: {"id": "XXXXX", "title": "فارسی", "is_selected": false}
        items = re.findall(
            r'"id"\s*:\s*"([A-Za-z0-9]{5,10})"\s*,\s*"title"\s*:\s*"([\u0600-\u06FF\s\u200c،\-()&]+)"',
            txt
        )
        for cat_id, cat_name in items:
            cat_name = cat_name.strip()
            if cat_name and cat_id not in all_cats:
                all_cats[cat_id] = cat_name
        
        print(f"[{query}] found {len(items)} categories, total now: {len(all_cats)}")
        
    except Exception as e:
        print(f"[{query}] ERROR: {e}")

print(f"\n\n{'='*60}")
print(f"TOTAL UNIQUE CATEGORIES FOUND: {len(all_cats)}")
print(f"{'='*60}")

# Sort by name
for cat_id, cat_name in sorted(all_cats.items(), key=lambda x: x[1]):
    print(f'  {{Key: "{cat_id}", Name: "{cat_name}"}},')

# Also print as Go struct format
print(f"\n\n--- Go TargetConfigs format ---")
for cat_id, cat_name in sorted(all_cats.items(), key=lambda x: x[1]):
    print(f'  {{Key: "{cat_id}", Name: "{cat_name}", SnappCat: "{cat_id}", DkQuery: ""}},')
