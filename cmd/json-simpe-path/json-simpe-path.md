% JSON-SIMPE-PATH(1) go2man
% Auto-generated
% Jun 2025

# NAME

json-simpe-path - query JSON files using lookup paths

# SYNOPSIS

`json-simpe-path [options] PATH [PATH ...]`

# DESCRIPTION

Reads one or more JSON documents and extracts values with lookup's SimplePath syntax. It mirrors yaml-simpe-path but defaults to JSON input and output.

# OPTIONS

See README for details.

# EXAMPLES

```
$ cat <<'J' > doc.json
{"name":"foo","spec":{"replicas":3},"metadata":{"name":"prod-service"}}
J
$ json-simpe-path -f doc.json .spec.replicas
3
```

```
$ json-simpe-path -grep '^prod' .metadata.name doc.json
prod-service
```

```
$ json-simpe-path -count .metadata.name doc.json
1
```

# SEE ALSO

yaml-simpe-path(1)
