# yaml-simple-path

yaml-simple-path is a small command line tool built on top of the lookup library. It reads one or more YAML documents and extracts values using lookup's `SimplePath` syntax. The interface should feel familiar to Unix users with options inspired by `cut`, `sed`, `grep` and modern tools such as `jq`.

```
Usage: yaml-simple-path [options] PATH [PATH ...]

Options:
  -f string   YAML file to read (default stdin)
  -e string   simple path query (can be repeated; processes multiple queries before positional paths)
  -d string   output delimiter (default "\n")
  -json       output as JSON
  -yaml       output as YAML (default)
  -raw        output raw value without formatting
  -grep str   only print results matching the regex
  -v          invert regex match
  -n          prefix results with their index
  -0          use NUL as output delimiter
  -count      only print the number of matched results
  -strict     strict mode: exit non-zero for invalid queries, missing paths, evaluation errors, or conflicting output flags
```

The tool expects one or more lookup paths. If `-e` is supplied, each occurrence is added in order as a query, followed by any additional positional query paths from the command line. Each YAML document in the input stream is decoded in turn and every query is executed against it.

Examples:

```bash
# Extract a field from a file
$ cat <<'EOF' > doc.yaml
name: foo
spec:
  replicas: 3
metadata:
  name: prod-service
EOF
$ yaml-simple-path -f doc.yaml .spec.replicas
3

# Output raw scalar values
$ yaml-simple-path -raw .spec.replicas < doc.yaml
3

# Only show values matching a pattern
$ yaml-simple-path -grep '^prod' -f doc.yaml .metadata.name
prod-service

# Count how many documents have a metadata.name field
$ yaml-simple-path -count -f doc.yaml .metadata.name
1
```

Multiple documents are separated by the chosen delimiter. By default results are printed as YAML but `-json` or `-raw` can be used for alternative output formats.
