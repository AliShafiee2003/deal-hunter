import urllib.request, urllib.error, json, time
exec(open('probe_snapp2.py').read().split("base, meta")[0])
base, meta = call({'page': 1})
print('base pagination', meta.get('pagination'))
for k in ['current_page', 'page_number', 'p', 'from', 'start', 'pageNumber', 'page_index']:
    for v in (2, 100):
        time.sleep(1.2)
        ids, m = call({k: v})
        if isinstance(ids, str):
            print(k, v, ids); continue
        print(k, v, 'n=%d overlap=%d cur=%s' % (len(ids), len(set(ids) & set(base)), (m or {}).get('pagination', {}).get('current_page')))
# page as string / large
for v in (3, 10):
    time.sleep(1.2)
    ids, m = call({'page': v})
    print('page', v, 'overlap=%d' % len(set(ids) & set(base)), m.get('pagination'))
