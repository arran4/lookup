import urllib.request
from html.parser import HTMLParser
import re

class TomlParser(HTMLParser):
    def __init__(self):
        super().__init__()
        self.in_code = False
        self.code_blocks = []
        self.current_block = []

    def handle_starttag(self, tag, attrs):
        if tag == 'code':
            for attr in attrs:
                if attr[0] == 'class' and 'language-toml' in attr[1]:
                    self.in_code = True

    def handle_endtag(self, tag):
        if tag == 'code' and self.in_code:
            self.in_code = False
            self.code_blocks.append(''.join(self.current_block))
            self.current_block = []

    def handle_data(self, data):
        if self.in_code:
            self.current_block.append(data)

url = 'https://arran4.github.io/blog/post/2026/006-github-ci-and-deploy/'
req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
html = urllib.request.urlopen(req).read().decode('utf-8')

html = re.sub(r'<span class=\"ln\">.*?</span>', '', html)

parser = TomlParser()
parser.feed(html)

for i, block in enumerate(parser.code_blocks):
    print(f"--- TOML BLOCK {i} ---")
    print(block)
