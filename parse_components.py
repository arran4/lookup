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
        m = re.match(r'^\s*\d+\s?(.*)$', line)
        if m:
            cleaned_lines.append(m.group(1))
        else:
            cleaned_lines.append(line)

    out_blocks.append('\n'.join(cleaned_lines))

with open('components.yml', 'w') as f:
    for i, block in enumerate(out_blocks):
        f.write(f'# BLOCK {i}\n')
        f.write(block)
        f.write('\n\n')
