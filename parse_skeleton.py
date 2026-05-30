import urllib.request
from html.parser import HTMLParser
import re

class SkeletonParser(HTMLParser):
    def __init__(self):
        super().__init__()
        self.in_skeleton_header = False
        self.in_code = False
        self.code_blocks = []
        self.current_block = []

    def handle_starttag(self, tag, attrs):
        if tag == 'h2':
            for attr in attrs:
                if attr[0] == 'id' and attr[1] == 'step-16-full-skeleton-compact-but-wired':
                    self.in_skeleton_header = True
        elif tag == 'code' and self.in_skeleton_header:
            for attr in attrs:
                if attr[0] == 'class' and 'language-yaml' in attr[1]:
                    self.in_code = True

    def handle_endtag(self, tag):
        if tag == 'code' and self.in_code:
            self.in_code = False
            self.in_skeleton_header = False # Stop after first code block under skeleton
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

parser = SkeletonParser()
parser.feed(html)

with open('skeleton.yml', 'w') as f:
    if parser.code_blocks:
        f.write(parser.code_blocks[0])
