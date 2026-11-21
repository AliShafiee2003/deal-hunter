import urllib.request, json, re, time

headers = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    'Content-Type': 'application/json',
    'Accept': 'application/json',
    'Origin': 'https://snappshop.ir',
    'Referer': 'https://snappshop.ir/'
}

time.sleep(3)
body = json.dumps({
    'query': 'موبایل', 'sort': '5DzQq4', 'is_available': True,
    'has_discount': True, 'page': 1, 'limit': 1
}).encode()
req = urllib.request.Request(
    'https://apix.snappshop.ir/search/v1?lat=35.77331&lng=51.418591',
    data=body, headers=headers
)
with urllib.request.urlopen(req, timeout=10) as r:
    data = json.loads(r.read())

txt = json.dumps(data, ensure_ascii=False)

# Find the category_tree filter block
filter_match = re.search(r'"filter_id"\s*:\s*"category_tree".*?"items"\s*:\s*(\[.*?\])', txt, re.DOTALL)
if filter_match:
    items_json = filter_match.group(1)
    items = json.loads(items_json)
    print(f"Category tree items: {len(items)}")
    for it in items:
        print(f"  {it.get('id')}: {it.get('title')}")
else:
    print("category_tree not found, searching for all id+title pairs...")
    items = re.findall(
        r'"id"\s*:\s*"([A-Za-z0-9]{5,10})"\s*,\s*"title"\s*:\s*"([\u0600-\u06FF\s\u200c\u060c\(\)&\-]+)"',
        txt
    )
    print(f"Total pairs: {len(items)}")
    for id_, name in items[:80]:
        print(f"  {id_}: {name}")

# Save full response for inspection
with open('snapp_response.json', 'w', encoding='utf-8') as f:
    json.dump(data, f, ensure_ascii=False, indent=2)
print("\nSaved full response to snapp_response.json")
print("Top-level keys:", list(data.keys()))
print("data keys:", list(data.get('data', {}).keys()))
structure = data.get('data', {}).get('structure', [])
for s in structure:
    print(f"  section: id={s.get('id')} type={s.get('section_type')} items={len(s.get('items',[]))} filter_id={s.get('filter_id')}")
