import re
import urllib.request
from html.parser import HTMLParser

class FinalParser(HTMLParser):
    def __init__(self):
        super().__init__()
        self.in_code = False
        self.code_blocks = []
        self.current_block = []

    def handle_starttag(self, tag, attrs):
        if tag == 'code':
            for attr in attrs:
                if attr[0] == 'class' and 'language-yaml' in attr[1]:
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

# Remove line numbers from HTML before parsing
html = re.sub(r'<span class=\"ln\">.*?</span>', '', html)
html = re.sub(r'<span class=\"cl\">', '', html)

parser = FinalParser()
parser.feed(html)

def clean_block(block):
    lines = block.split('\n')
    cleaned = []
    for line in lines:
        m = re.match(r'^\s*\d+\s?(.*)$', line)
        if m:
            cleaned.append(m.group(1))
        else:
            cleaned.append(line)
    return '\n'.join(cleaned).strip()

blocks = [clean_block(b) for b in parser.code_blocks]

# Skeleton is block 37 (index 37, assuming 0-indexed. Let's just output all to a python dict for easy access)
with open('assemble.py', 'w') as f:
    f.write('blocks = {\n')
    for i, b in enumerate(blocks):
        f.write(f'  {i}: """{b}""",\n')
    f.write('}\n')
