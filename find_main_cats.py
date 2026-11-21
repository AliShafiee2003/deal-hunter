"""
Get main SnappShop categories by querying broad terms and extracting category_tree filter.
This gives us MAIN category codes (not brands/colors/etc).
"""
import urllib.request, json, re, time

headers = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
    'Content-Type': 'application/json',
    'Accept': 'application/json',
    'Origin': 'https://snappshop.ir',
    'Referer': 'https://snappshop.ir/'
}
BASE = 'https://apix.snappshop.ir/search/v1?lat=35.77331&lng=51.418591'

# Map: main category name -> what to search for it
CATEGORY_SEARCHES = {
    'موبایل': 'گوشی موبایل سامسونگ',
    'لپ تاپ': 'لپ تاپ ایسر',
    'هارد اکسترنال': 'هارد اکسترنال سیگیت',
    'SSD': 'اس اس دی سامسونگ',
    'هدفون': 'هدفون بلوتوث',
    'اسپیکر': 'اسپیکر بلوتوث',
    'ساعت هوشمند': 'ساعت هوشمند شیائومی',
    'پاوربانک': 'پاوربانک انکر',
    'دوربین': 'دوربین کانن',
    'تبلت': 'تبلت سامسونگ',
    'لوازم خانگی': 'ماشین لباسشویی',
    'یخچال': 'یخچال ال جی',
    'جارو برقی': 'جارو برقی فیلیپس',
    'آرایشی': 'کرم پودر آرایشی',
    'مکمل سلامت': 'مکمل پروتئین',
    'عینک': 'عینک آفتابی',
    'پوشاک': 'تیشرت مردانه',
    'کفش': 'کفش ورزشی',
    'طلا': 'انگشتر طلا',
    'کتاب': 'کتاب رمان',
    'ابزار': 'دریل بوش',
    'خودرو': 'روغن موتور',
    'اسباب بازی': 'لگو اسباب بازی',
    'ورزشی': 'دمبل ورزشی',
    'سوپرمارکت': 'آبمیوه',
    'قهوه ساز': 'قهوه ساز دلونگی',
    'مبل': 'مبل راحتی',
    'کنسول': 'پلی استیشن',
    'گیمینگ': 'ماوس گیمینگ',
}

results = {}  # category_name -> (id, title)

for cat_name, query in CATEGORY_SEARCHES.items():
    time.sleep(2)
    body = json.dumps({
        'query': query, 'sort': '5DzQq4', 'is_available': True,
        'has_discount': True, 'page': 1, 'limit': 1
    }).encode()
    try:
        req = urllib.request.Request(BASE, data=body, headers=headers)
        with urllib.request.urlopen(req, timeout=10) as r:
            data = json.loads(r.read())
        
        # Find category_tree
        for section in data.get('data', {}).get('structure', []):
            for filt in section.get('filter', []):
                if filt.get('filter_id') == 'category_tree':
                    items = filt.get('items', [])
                    print(f"\n[{cat_name}] query='{query}' => {len(items)} categories:")
                    for it in items:
                        print(f"  {it['id']}: {it['title']}")
                        if cat_name not in results:
                            # First category is usually the main one
                            results[cat_name] = (it['id'], it['title'])
                    break
    except Exception as e:
        print(f"[{cat_name}] ERROR: {e}")

print("\n\n" + "="*60)
print("MAIN CATEGORY CODES MAPPING:")
print("="*60)
for cat_name, (cat_id, cat_title) in results.items():
    print(f"  {cat_name} -> {cat_id} ({cat_title})")

print("\n\nGo SnappCat fields:")
# Map to our keys
key_map = {
    'موبایل': 'mobile',
    'لپ تاپ': 'laptop',
    'هارد اکسترنال': 'hard',
    'SSD': 'ssd',
    'هدفون': 'headphone',
    'اسپیکر': 'speaker',
    'ساعت هوشمند': 'gadgets',
    'پاوربانک': 'powerbank',
    'دوربین': 'camera',
    'تبلت': 'tablet',
    'لوازم خانگی': 'home',
    'یخچال': 'fridge',
    'جارو برقی': 'vacuum',
    'آرایشی': 'beauty',
    'مکمل سلامت': 'supplement',
    'عینک': 'eyewear',
    'پوشاک': 'fashion',
    'کفش': 'shoes',
    'طلا': 'gold',
    'کتاب': 'book_art',
    'ابزار': 'tools',
    'خودرو': 'auto',
    'اسباب بازی': 'toy',
    'ورزشی': 'sport',
    'سوپرمارکت': 'supermarket',
    'قهوه ساز': 'coffee',
    'مبل': 'furniture',
    'کنسول': 'gaming',
    'گیمینگ': 'gaming_gear',
}
for cat_name, (cat_id, cat_title) in results.items():
    key = key_map.get(cat_name, cat_name)
    print(f'  {key}: SnappCat="{cat_id}"  // {cat_title}')
