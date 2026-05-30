import urllib.request
import re
from html import unescape

url = 'https://arran4.github.io/blog/post/2026/006-github-ci-and-deploy/'
req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
html = urllib.request.urlopen(req).read().decode('utf-8')

blocks = re.findall(r'<div class=\"highlight\"><pre.*?>.*?<code.*?>(.*?)</code></pre></div>', html, re.DOTALL)
for i, b in enumerate(blocks):
    b_clean = re.sub(r'<span class=\"ln\">.*?</span>', '', b)
    b_text = re.sub(r'<[^>]+>', '', b_clean)
    if 'gitleaks config' in b_text:
        print(unescape(b_text))
