import re

with open('all_blocks.txt', 'r') as f:
    text = f.read()

blocks = text.split('--- BLOCK ')
out_blocks = []

for block in blocks:
    if not block.strip():
        continue
    # remove the number and --- \n
    lines = block.split('\n')[1:]

    # Strip line numbers from the start of lines
    cleaned_lines = []
    for line in lines:
        # Match optional spaces, then digits, then a single space or just digits
        # The line numbers are formatted like `  1` or ` 10` followed by the code
        m = re.match(r'^\s*\d+\s?(.*)$', line)
        if m:
            cleaned_lines.append(m.group(1))
        else:
            cleaned_lines.append(line)

    out_blocks.append('\n'.join(cleaned_lines))

with open('full_template.yml', 'w') as f:
    # write the skeleton block (block 37 is the skeleton)
    f.write(out_blocks[37])
