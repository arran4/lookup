import re

with open('all_blocks.txt', 'r') as f:
    text = f.read()

raw_blocks = text.split('--- BLOCK ')
out_blocks = []

for block in raw_blocks:
    if not block.strip():
        continue
    lines = block.split('\n')[1:]

    cleaned_lines = []
    for line in lines:
        m = re.match(r'^\s*\d+\s?(.*)$', line)
        if m:
            cleaned_lines.append(m.group(1))
        else:
            cleaned_lines.append(line)

    out_blocks.append('\n'.join(cleaned_lines).rstrip())

print(out_blocks[8])
print("-----")
print(out_blocks[9])
