from assemble import blocks

# Blocks we need:
# 0: comment
# 1: triggers
# 2: prepare-release-tag
# 3: route job
# 5: discover
# 7: gitleaks, dependency-review
# 11: golangci, go-test, go-vet, go-fmt-pr
# 20: autofix
# 21: cleanup-autofix-prs
# 24: goreleaser
# 35: publish-draft, promote-release
# 36: prepare-next-version-pr

# Wait, go-test in block 11 has a matrix and caching which is different from the project's existing test.yml
# Actually, the guide says: "Use this baseline...". And we should merge the existing `test.yml` into it.
