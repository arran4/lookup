import urllib.request
from bs4 import BeautifulSoup
import re

url = 'https://arran4.github.io/blog/post/2026/006-github-ci-and-deploy/'
req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
html = urllib.request.urlopen(req).read().decode('utf-8')

# Let's extract without BS4
blocks = re.findall(r'<div class=\"highlight\"><pre.*?>.*?<code.*?>(.*?)</code></pre></div>', html, re.DOTALL)
for i, b in enumerate(blocks):
    b_text = re.sub(r'<[^>]+>', '', b)
    if 'gitleaks config' in b_text:
        print(f"Found in block {i}:")
        print(b_text)
