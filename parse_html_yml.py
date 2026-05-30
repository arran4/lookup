from bs4 import BeautifulSoup
import sys

def extract_code(html_file):
    with open(html_file, 'r') as f:
        soup = BeautifulSoup(f, 'html.parser')

    for code_block in soup.find_all('code', class_='language-yaml'):
        lines = []
        for line in code_block.find_all('span', class_='line'):
            # find the span class 'cl' which has the code string
            cl = line.find('span', class_='cl')
            if cl:
                line_text = ""
                # remove the span class 'ln' (line number) from cl's children if it exists
                for child in cl.children:
                    if child.name == 'span' and 'ln' in child.get('class', []):
                        continue
                    line_text += child.text
                lines.append(line_text)
            else:
                lines.append("")

        print("".join(lines))
        print("---")

extract_code(sys.argv[1])
