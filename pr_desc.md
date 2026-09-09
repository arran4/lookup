Part of #84

- **Corrected release metadata**: Updated `.goreleaser.yml` to specify `BSD-3-Clause` instead of `Private`, provided a descriptive `description` ("Command-line tool to query JSON and YAML files using lookup paths"), and set a valid `homepage` link to the GitHub repository.
- **Expected release artifact types**: The binaries (`json-simple-path`, `yaml-simple-path`, `json-simpe-path`, and `yaml-simpe-path`) are correctly compiled and bundled into the expected artifacts (`zip`, `tar.gz`, `deb`, `rpm`, `apk`, `archlinux`, `termux.deb`) and manpages are properly added to the release archives and nfpm packages. We also fixed GoReleaser v2 configuration deprecations (`archives.ids`, `formats`, `version_template`). Android target was added to support termux.deb package.
- **GoReleaser verification performed**: Added a `goreleaser-check` job to the `ci.yml` workflow that runs `goreleaser check` and `goreleaser release --snapshot --clean` on PRs and validates the produced artifacts (archives, binaries, checksums, and packages, including inspecting their inner contents like man pages).
- **CI before/after responsibility table**:
    - `go test`: Before: `ci.yml`, `test.yml`. After: `ci.yml`
    - `go vet`: Before: `ci.yml`, `test.yml`. After: `ci.yml`
    - `lint` (golangci-lint): Before: `golangci-lint.yml`. After: `ci.yml`
    - `Go-version selection`: Before: Multiple versions in `test.yml` and `golangci-lint.yml`. After: Centralized in `ci.yml` using `go-version-file: go.mod`.
    - `OS matrix`: Before: ubuntu/macos/windows in `test.yml`. After: ubuntu/macos/windows in `ci.yml`
    - `coverage`: Before: `test.yml` and `ci.yml`. After: `ci.yml`
    - `benchmarks`: Before: `test.yml` and `ci.yml` (creating duplicate PR comments). After: `ci.yml` (now using updating non-duplicative PR comments).
    - `build validation`: Before: `ci.yml`. After: `ci.yml`
    - `release validation`: Before: not tested via Goreleaser snapshot. After: `ci.yml` runs Goreleaser check with thorough artifact verification on PRs.
    - `release publication`: Before: `ci.yml`. After: `ci.yml`
- **Workflows removed/narrowed and why**: Removed `.github/workflows/test.yml` and `.github/workflows/golangci-lint.yml` as they are now fully handled and verified by the unified `.github/workflows/ci.yml`. This prevents duplicate compute and overlapping tests/lints.
- **Confirmation that release ownership/routing was not changed**: The explicit-dispatch single-release-owner architecture in `ci.yml` is strictly maintained. GoReleaser remains the sole GitHub Release owner, the explicit-dispatch/tag-context release routing from #83 was not changed, and no production release was created by these changes. Snapshot validation does NOT prove the historical missing-assets problem is completely fixed; that remains to be demonstrated by a later controlled release.
