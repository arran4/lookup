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

    out_blocks.append('\n'.join(cleaned_lines).strip())

# Identify blocks by some unique string in them:
for i, b in enumerate(out_blocks):
    if "Agent rules for generation" in b: print(f"{i} Agent rules")
    if "name: CI/CD" in b and "on:" in b and "push" in b: print(f"{i} Triggers")
    if "outputs:" in b and "run_code_checks" in b and "run_pr_meta_checks" in b: print(f"{i} Route")
    if "id: detect" in b and "EXPECT_GO" in b: print(f"{i} Discover")
    if "name: Secret scan" in b: print(f"{i} Security")
    if "name: lint" in b and "uses: golangci/golangci-lint-action" in b: print(f"{i} Golang")
    if "name: Auto-format and open PR" in b: print(f"{i} Autofix")
    if "name: Cleanup autofix PRs on parent close" in b: print(f"{i} Cleanup autofix")
    if "name: GoReleaser" in b: print(f"{i} GoReleaser")
    if "name: Publish draft release assets" in b: print(f"{i} Publish draft")
    if "name: Prepare next development iteration PR" in b: print(f"{i} Prepare next")
    if "name: Prepare release tag" in b: print(f"{i} Prepare release tag")
