import urllib.request, urllib.error, json, time
exec(open('probe_snapp2.py').read().split("base, meta")[0])
base, meta = call({'skip': 0})
print('skip0', len(base), meta['pagination'])
allids = set(base)
for sk in range(1, 12):
    time.sleep(1.5)
    ids, m = call({'skip': sk, 'exclude_filters': True})
    if isinstance(ids, str): print('skip', sk, ids); break
    new = len(set(ids) - allids); allids |= set(ids)
    print('skip', sk, 'n=%d new=%d cur=%s' % (len(ids), new, m['pagination']['current_page']))
print('unique', len(allids))
