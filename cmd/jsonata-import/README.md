# JSONata fixture importer

The importer reads the JSONata language-neutral test suite from the exact commit identified by `UpstreamRev` in `main.go` (JSONata v2.0.2). Run these commands from the repository root:

```sh
go run ./cmd/jsonata-import
go run ./cmd/jsonata-import verify
go test ./cmd/jsonata-import ./jsonata
```

Generation updates the managed `jsonata/testdata/test-suite/{groups,datasets}` files, removing stale entries. Verification checks the complete managed file set and its contents without modifying files. Review **all** generated fixture and baseline changes before committing. If the upstream source changes, investigate and document the expected-failure and compatibility-matrix deltas; do not equate corrected fixtures with improved runtime compatibility.

Importer unit tests construct a small compressed upstream-shaped archive and write exclusively into temporary directories. The network-dependent `verify` command is an integration check against the pinned upstream archive; it is not part of ordinary unit tests.
