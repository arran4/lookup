package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/arran4/lookup"
	"gopkg.in/yaml.v3"
)

func usage(fs *flag.FlagSet, defaultFormat string) {
	_, _ = fmt.Fprintf(fs.Output(), `Usage: %s [options] PATH [PATH ...]
Options:
  -f string  %s file to read (default stdin)
  -e string  simple path query (can be repeated)
  -d string  output delimiter (default "\n")
  -json      output as JSON
  -yaml      output as YAML
  -raw       output raw values without formatting
  -grep str  only print results matching the regex
  -v         invert grep match
  -n         prefix results with their index
  -0         use NUL as output delimiter
  -count     only print the number of matched results
  -strict    strict mode: exit non-zero for invalid queries, missing paths, or evaluation errors
`, fs.Name(), defaultFormat)
}

type stringSlice []string

func (i *stringSlice) String() string {
	return strings.Join(*i, " ")
}

func (i *stringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// Run executes the CLI logic for querying JSON or YAML files.
func Run(name string, args []string, stdin io.Reader, stdout, stderr io.Writer, defaultFormat string) error {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)

	file := fs.String("f", "", "input file")
	var queryFlags stringSlice
	fs.Var(&queryFlags, "e", "simple path query")
	delim := fs.String("d", "\n", "output delimiter")
	jsonOut := fs.Bool("json", false, "output JSON")
	yamlOut := fs.Bool("yaml", false, "output YAML")
	rawOut := fs.Bool("raw", false, "output raw values")
	grepExpr := fs.String("grep", "", "filter by regex")
	invert := fs.Bool("v", false, "invert regex match")
	number := fs.Bool("n", false, "number results")
	nullDelim := fs.Bool("0", false, "use NUL as delimiter")
	countOnly := fs.Bool("count", false, "only print match count")
	strictMode := fs.Bool("strict", false, "enable strict mode")

	fs.Usage = func() { usage(fs, defaultFormat) }
	if err := fs.Parse(args); err != nil {
		return err
	}

	queries := []string{}
	if len(queryFlags) > 0 {
		queries = append(queries, queryFlags...)
	}
	queries = append(queries, fs.Args()...)

	outFlags := 0
	if *jsonOut {
		outFlags++
	}
	if *yamlOut {
		outFlags++
	}
	if *rawOut {
		outFlags++
	}
	if *strictMode && outFlags > 1 {
		return fmt.Errorf("conflicting output flags: -json, -yaml, and -raw are mutually exclusive in strict mode")
	}
	if len(queries) == 0 {
		fs.Usage()
		return fmt.Errorf("no query provided")
	}

	var compiledQueries []*lookup.Relator
	for _, q := range queries {
		if *strictMode {
			rel, err := lookup.CompileSimplePath(q)
			if err != nil {
				return fmt.Errorf("compile %q: %w", q, err)
			}
			compiledQueries = append(compiledQueries, rel)
		} else {
			rel := lookup.ParseSimplePath(q)
			compiledQueries = append(compiledQueries, rel)
		}
	}

	if *nullDelim {
		*delim = "\x00"
	}

	r := stdin
	if *file != "" {
		f, err := os.Open(*file)
		if err != nil {
			return fmt.Errorf("open %s: %w", *file, err)
		}
		defer func() {
			_ = f.Close()
		}()
		r = f
	}

	var dec interface {
		Decode(v interface{}) error
	}
	if defaultFormat == "YAML" {
		dec = yaml.NewDecoder(r)
	} else {
		dec = json.NewDecoder(r)
	}

	var re *regexp.Regexp
	var err error
	if *grepExpr != "" {
		re, err = regexp.Compile(*grepExpr)
		if err != nil {
			return fmt.Errorf("invalid regex: %w", err)
		}
	}

	index := 0
	count := 0
	first := true
	for {
		var doc interface{}
		err := dec.Decode(&doc)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("decode: %w", err)
		}
		for _, q := range compiledQueries {
			root := lookup.Reflect(doc)
			res := q.Run(lookup.NewScope(nil, root))

			if res == nil {
				if *strictMode {
					return fmt.Errorf("nil result")
				}
				continue
			}

			if _, isInvalidor := res.(*lookup.Invalidor); isInvalidor {
				if *strictMode {
					// Extract the underlying error if possible
					if err, ok := res.(error); ok {
						return fmt.Errorf("evaluation error: %w", err)
					}
					return fmt.Errorf("missing path or evaluation error")
				}
				continue
			}

			val := res.Raw()
			if re != nil {
				matched := re.MatchString(fmt.Sprint(val))
				if *invert {
					matched = !matched
				}
				if !matched {
					continue
				}
			}
			count++
			if *countOnly {
				continue
			}
			if !first {
				_, _ = fmt.Fprint(stdout, *delim)
			}
			first = false
			if *number {
				_, _ = fmt.Fprintf(stdout, "%d:", index)
			}
			if *rawOut {
				_, _ = fmt.Fprint(stdout, fmt.Sprint(val))
			} else if defaultFormat == "YAML" {
				if *jsonOut {
					b, err := json.Marshal(val)
					if err != nil {
						return fmt.Errorf("json encode: %w", err)
					}
					_, _ = fmt.Fprint(stdout, string(b))
				} else {
					b, err := yaml.Marshal(val)
					if err != nil {
						return fmt.Errorf("yaml encode: %w", err)
					}
					_, _ = fmt.Fprint(stdout, strings.TrimSuffix(string(b), "\n"))
				}
			} else { // JSON default format
				if *yamlOut {
					b, err := yaml.Marshal(val)
					if err != nil {
						return fmt.Errorf("yaml encode: %w", err)
					}
					_, _ = fmt.Fprint(stdout, strings.TrimSuffix(string(b), "\n"))
				} else {
					b, err := json.Marshal(val)
					if err != nil {
						return fmt.Errorf("json encode: %w", err)
					}
					_, _ = fmt.Fprint(stdout, string(b))
				}
			}
			index++
		}
	}
	if *countOnly {
		_, _ = fmt.Fprint(stdout, count)
	}
	return nil
}
