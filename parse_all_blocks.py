import urllib.request
from html.parser import HTMLParser
import re

class AllBlocksParser(HTMLParser):
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

# Remove span line numbers
html = re.sub(r'<span class=\"ln\">.*?</span>', '', html)
# also remove `<span class="cl">` which is added
html = re.sub(r'<span class=\"cl\">', '', html)

parser = AllBlocksParser()
parser.feed(html)

with open('all_blocks.txt', 'w') as f:
    for i, block in enumerate(parser.code_blocks):
        f.write(f'--- BLOCK {i} ---\n')
        # strip the line numbers that might be left over as spaces
        lines = [line[line.find(line.lstrip()):] if line.lstrip() else "" for line in block.split('\n')]

        # some lines start with spaces for indentation. we shouldn't strip everything.
        # let's just write the block.
        f.write(block)
        f.write('\n\n')
