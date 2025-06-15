package main

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

func usage(fs *flag.FlagSet) {
	fmt.Fprintf(fs.Output(), `Usage: %s [options] PATH [PATH ...]
Options:
  -f string  YAML file to read (default stdin)
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
`, fs.Name())
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("yaml-simpe-path", flag.ContinueOnError)
	fs.SetOutput(stderr)

	file := fs.String("f", "", "input file")
	queryFlag := fs.String("e", "", "simple path query")
	delim := fs.String("d", "\n", "output delimiter")
	jsonOut := fs.Bool("json", false, "output JSON")
	yamlOut := fs.Bool("yaml", false, "output YAML")
	rawOut := fs.Bool("raw", false, "output raw values")
	grepExpr := fs.String("grep", "", "filter by regex")
	invert := fs.Bool("v", false, "invert regex match")
	number := fs.Bool("n", false, "number results")
	nullDelim := fs.Bool("0", false, "use NUL as delimiter")
	countOnly := fs.Bool("count", false, "only print match count")
	fs.Usage = func() { usage(fs) }
	if err := fs.Parse(args); err != nil {
		return err
	}

	queries := []string{}
	if *queryFlag != "" {
		queries = append(queries, *queryFlag)
	}
	queries = append(queries, fs.Args()...)
	if len(queries) == 0 {
		fs.Usage()
		return fmt.Errorf("no query provided")
	}

	if *nullDelim {
		*delim = "\x00"
	}

	var r io.Reader = stdin
	if *file != "" {
		f, err := os.Open(*file)
		if err != nil {
			return fmt.Errorf("open %s: %w", *file, err)
		}
		defer f.Close()
		r = f
	}

	dec := yaml.NewDecoder(r)

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
		for _, q := range queries {
			res := lookup.QuerySimplePath(doc, q)
			if res == nil {
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
				fmt.Fprint(stdout, *delim)
			}
			first = false
			if *number {
				fmt.Fprintf(stdout, "%d:", index)
			}
			switch {
			case *rawOut:
				fmt.Fprint(stdout, fmt.Sprint(val))
			case *jsonOut:
				b, err := json.Marshal(val)
				if err != nil {
					return fmt.Errorf("json encode: %w", err)
				}
				fmt.Fprint(stdout, string(b))
			case *yamlOut || (!*jsonOut && !*rawOut):
				b, err := yaml.Marshal(val)
				if err != nil {
					return fmt.Errorf("yaml encode: %w", err)
				}
				fmt.Fprint(stdout, strings.TrimSuffix(string(b), "\n"))
			}
			index++
		}
	}
	if *countOnly {
		fmt.Fprint(stdout, count)
	}
	return nil
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
