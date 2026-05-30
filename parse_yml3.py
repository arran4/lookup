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

print("BLOCK 4:")
print(out_blocks[4])
print("-----")
print("BLOCK 5:")
print(out_blocks[5])
print("-----")
print("BLOCK 6:")
print(out_blocks[6])
print("-----")
print("BLOCK 7:")
print(out_blocks[7])
print("-----")
print("BLOCK 8:")
print(out_blocks[8])
print("-----")
