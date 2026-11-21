import urllib.request, urllib.error, json, time
exec(open('probe_snapp2.py').read().split("base, meta")[0])
base, meta = call({})
print('base', meta['pagination'])
for ue in ['&page=2', '&page=1', '&current_page=2', '&p=2', '&pagination[current_page]=2', '&page_number=2']:
    time.sleep(1.2)
    ids, m = call({}, ue)
    if isinstance(ids, str): print(ue, ids); continue
    print(ue, 'overlap=%d' % len(set(ids) & set(base)), m['pagination']['current_page'])
# body nested
for body in [{'pagination': {'page': 2}}, {'pagination': {'current_page': 2}}, {'paging': {'page': 2}}]:
    time.sleep(1.2)
    ids, m = call(body)
    if isinstance(ids, str): print(body, ids); continue
    print(body, 'overlap=%d' % len(set(ids) & set(base)), m['pagination']['current_page'])
