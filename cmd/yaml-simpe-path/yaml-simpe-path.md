% YAML-SIMPE-PATH(1) go2man
% Auto-generated
% Jun 2025

# NAME

yaml-simpe-path - query YAML files using lookup paths

# SYNOPSIS

`yaml-simpe-path [options] PATH [PATH ...]`

# DESCRIPTION

Reads one or more YAML documents and extracts values with lookup's SimplePath syntax. It mimics unix tools like cut, sed and grep while adding jq-like features.

# OPTIONS

See README for details.

# EXAMPLES

```
$ cat <<'Y' > doc.yaml
name: foo
spec:
  replicas: 3
metadata:
  name: prod-service
Y
$ yaml-simpe-path -f doc.yaml .spec.replicas
3
```

```
$ yaml-simpe-path -grep '^prod' .metadata.name doc.yaml
prod-service
```

```
$ yaml-simpe-path -count .metadata.name doc.yaml
1
```

# SEE ALSO

json-simpe-path(1)
