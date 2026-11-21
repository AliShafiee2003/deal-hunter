"""
Extract SnappShop category codes from their JS bundles (no API calls needed).
Category codes are embedded in the Next.js JS chunks.
"""
import urllib.request, re, json, time

headers = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
    'Accept': '*/*',
    'Referer': 'https://snappshop.ir/'
}

# Step 1: get the main page to find JS chunk URLs
print("Fetching homepage HTML...")
req = urllib.request.Request('https://snappshop.ir/', headers={**headers, 'Accept': 'text/html'})
with urllib.request.urlopen(req, timeout=12) as r:
    html = r.read().decode('utf-8', errors='replace')

# Find JS chunk URLs
js_chunks = re.findall(r'"([^"]*/_next/static/chunks/[^"]+\.js)"', html)
js_chunks += re.findall(r'src="([^"]*/_next/static/[^"]+\.js)"', html)
js_chunks = list(dict.fromkeys(js_chunks))
print(f"Found {len(js_chunks)} JS chunks")

# Step 2: search chunks for category codes
# SnappShop category codes are ~6 char alphanumeric like AqZkDJ, 50a6Zq
# They appear near Persian category names in JS

all_cats = {}

for chunk_url in js_chunks[:20]:  # check first 20 chunks
    if not chunk_url.startswith('http'):
        chunk_url = 'https://snappshop.ir' + chunk_url
    
    try:
        time.sleep(0.3)
        req = urllib.request.Request(chunk_url, headers=headers)
        with urllib.request.urlopen(req, timeout=10) as r:
            js = r.read().decode('utf-8', errors='replace')
        
        # Look for patterns like: "id":"AqZkDJ","name":"موبایل"
        # or: category:"AqZkDJ",name:"موبایل"
        # Known codes: AqZkDJ=mobile, w02XDY=laptop, 50a6Zq=hard, nqPlzq=ssd
        
        # Pattern 1: JSON-like objects with id+name
        pairs1 = re.findall(r'["\']id["\']\s*:\s*["\']([A-Za-z0-9]{5,8})["\'].*?["\'](?:name|title)["\']:\s*["\']([\u0600-\u06FF\s\u200c&()،]+)["\']', js)
        # Pattern 2: category slug patterns
        pairs2 = re.findall(r'slug["\']?\s*:\s*["\']([A-Za-z0-9]{5,8})["\'].*?name["\']?\s*:\s*["\']([\u0600-\u06FF\s\u200c&()،]+)["\']', js)
        
        for code, name in pairs1 + pairs2:
            name = name.strip()
            if name and len(name) > 1:
                all_cats[code] = name
                
        if pairs1 or pairs2:
            print(f"  Found in {chunk_url.split('/')[-1]}: {len(pairs1+pairs2)} entries")
            
    except Exception as e:
        pass

print(f"\n=== Found {len(all_cats)} category codes ===")
for code, name in sorted(all_cats.items(), key=lambda x: x[1]):
    print(f"  {code}: {name}")

# Also search for the category menu structure specifically
print("\n--- Searching for menu/nav structure ---")
for chunk_url in js_chunks[:20]:
    if not chunk_url.startswith('http'):
        chunk_url = 'https://snappshop.ir' + chunk_url
    try:
        req = urllib.request.Request(chunk_url, headers=headers)
        with urllib.request.urlopen(req, timeout=10) as r:
            js = r.read().decode('utf-8', errors='replace')
        # search for known code AqZkDJ to find the right chunk
        if 'AqZkDJ' in js or '50a6Zq' in js or 'w02XDY' in js:
            print(f"\nFOUND KEY CHUNK: {chunk_url.split('/')[-1]} (len={len(js)})")
            # Extract context around known codes
            for known in ['AqZkDJ', '50a6Zq', 'w02XDY', 'nqPlzq', '5Db1Lq', 'KqJkXD', 'ngywqR']:
                idx = js.find(known)
                if idx > 0:
                    snippet = js[max(0,idx-200):idx+200]
                    print(f"\n  [{known}] context: ...{snippet}...")
            break
    except:
        pass
