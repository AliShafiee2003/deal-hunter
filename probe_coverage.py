import urllib.request, urllib.error, json, time
exec(open('probe_snapp2.py').read().split("base, meta")[0])

ours = json.loads(urllib.request.urlopen('http://localhost:8080/api/deals?store=snapp&category=all&min_discount=1').read())
ours_ids = {d['id'][4:] for d in ours}
print('ours snapp deals:', len(ours_ids))

def probe(label, extra):
    ids = set()
    for sk in range(0, 5):
        time.sleep(1.6)
        r, m = call(dict(extra, skip=sk))
        if isinstance(r, str):
            print(label, 'skip', sk, r); break
        ids |= set(r)
        if len(r) < 50: break
    miss = ids - ours_ids
    print('%-22s seen=%4d  in_ours=%4d  missing=%4d  coverage=%.0f%%' % (label, len(ids), len(ids) - len(miss), len(miss), 100.0 * (len(ids) - len(miss)) / max(1, len(ids))))

probe('unfiltered', {})
for q in ['هارد اکسترنال', 'گوشی', 'کفش', 'لباس', 'کتاب', 'اسباب بازی', 'لوازم آشپزخانه', 'کرم', 'ابزار', 'مبل']:
    probe('q=' + q, {'query': q})
