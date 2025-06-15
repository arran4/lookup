# json-simpe-path

json-simpe-path is a small command line tool built on top of the lookup library. It reads one or more JSON documents and extracts values using lookup's `SimplePath` syntax. The interface should feel familiar to Unix users with options inspired by `cut`, `sed`, `grep` and modern tools such as `jq`.

```
Usage: json-simpe-path [options] PATH [PATH ...]

Options:
  -f string   JSON file to read (default stdin)
  -e string   simple path query (can be repeated)
  -d string   output delimiter (default "\n")
  -json       output as JSON
  -yaml       output as YAML
  -raw        output raw value without formatting
  -grep str   only print results matching the regex
  -v          invert regex match
  -n          prefix results with their index
  -0          use NUL as output delimiter
  -count      only print the number of matched results
```

The tool expects one or more lookup paths. If `-e` is supplied the flag value is treated as the first query followed by any additional paths on the command line. Each JSON document in the input stream is decoded in turn and every query is executed against it.

Examples:

```bash
# Extract a field from a file
$ cat <<'EOF' > doc.json
{"name":"foo","spec":{"replicas":3},"metadata":{"name":"prod-service"}}
EOF
$ json-simpe-path -f doc.json .spec.replicas
3

# Output raw scalar values
$ json-simpe-path -raw .spec.replicas < doc.json
3

# Only show values matching a pattern
$ json-simpe-path -grep '^prod' -raw -f doc.json .metadata.name
prod-service

# Count how many documents have a metadata.name field
$ json-simpe-path -count -f doc.json .metadata.name
1
```

Multiple documents are separated by the chosen delimiter. By default results are printed as JSON but `-json` or `-raw` can be used for alternative output formats.
