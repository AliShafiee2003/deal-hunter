"""
Discover SnappShop category codes by calling their categories API
with retry & backoff. Server is paused so no rate limit conflict.
"""
import urllib.request, json, time, random

headers = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
    'Accept': 'application/json',
    'Origin': 'https://snappshop.ir',
    'Referer': 'https://snappshop.ir/'
}

LAT = "35.77331"
LNG = "51.418591"

def call_api(url, body=None, retries=3):
    for attempt in range(retries):
        try:
            if attempt > 0:
                wait = 3 * (2 ** attempt) + random.uniform(1, 3)
                print(f"  Retry {attempt+1} after {wait:.1f}s...")
                time.sleep(wait)
            
            if body:
                data = json.dumps(body).encode()
                req = urllib.request.Request(url, data=data, headers={**headers, 'Content-Type': 'application/json'})
            else:
                req = urllib.request.Request(url, headers=headers)
            
            with urllib.request.urlopen(req, timeout=12) as r:
                return json.loads(r.read())
        except urllib.error.HTTPError as e:
            if e.code == 429:
                print(f"  429 rate limit on attempt {attempt+1}")
                if attempt == retries - 1:
                    raise
            else:
                raise
        except Exception as e:
            if attempt == retries - 1:
                raise
    return None

# Wait for rate limit to clear
print("Waiting 10s for rate limit to clear...")
time.sleep(10)

# Try the categories endpoint
cat_endpoints = [
    f"https://apix.snappshop.ir/catalog/v1/categories?lat={LAT}&lng={LNG}",
    f"https://apix.snappshop.ir/catalog/v2/categories?lat={LAT}&lng={LNG}",
    f"https://apix.snappshop.ir/vendor/v1/categories?lat={LAT}&lng={LNG}",
    f"https://apix.snappshop.ir/catalog/v1/landing?lat={LAT}&lng={LNG}",
]

for ep in cat_endpoints:
    print(f"\nTrying: {ep}")
    try:
        data = call_api(ep)
        print(f"SUCCESS! Keys: {list(data.keys()) if isinstance(data, dict) else type(data)}")
        txt = json.dumps(data, ensure_ascii=False)
        # Look for category ids (6-8 char alphanumeric)
        import re
        ids = re.findall(r'"(?:id|slug|code)"\s*:\s*"([A-Za-z0-9]{5,10})"', txt)
        names = re.findall(r'"(?:name|title|label)"\s*:\s*"([\u0600-\u06FF\s\u200c&()،0-9]+)"', txt)
        print(f"Potential IDs: {ids[:20]}")
        print(f"Persian names: {names[:20]}")
        print(json.dumps(data, ensure_ascii=False, indent=2)[:3000])
        break
    except Exception as e:
        print(f"  FAIL: {e}")

# Also try the search API to get category filter options
print("\n\nTrying search with filter discovery...")
time.sleep(5)
try:
    body = {
        "query": "موبایل",
        "sort": "5DzQq4",
        "is_available": True,
        "has_discount": True,
        "page": 1,
        "limit": 1
    }
    data = call_api(f"https://apix.snappshop.ir/search/v1?lat={LAT}&lng={LNG}", body)
    # Look for filter/facet data that might contain categories
    txt = json.dumps(data, ensure_ascii=False)
    import re
    # Find any section that looks like category filters
    filter_sections = re.findall(r'"(?:filter|facet|category|categories)":\s*(\[.*?\])', txt[:10000], re.DOTALL)
    print(f"Filter sections found: {len(filter_sections)}")
    for fs in filter_sections[:3]:
        print(fs[:500])
    
    # Print full structure keys
    if isinstance(data, dict):
        print("\nTop-level keys:", list(data.keys()))
        d = data.get('data', {})
        print("data keys:", list(d.keys()) if isinstance(d, dict) else type(d))
        
except Exception as e:
    print(f"Search API failed: {e}")
