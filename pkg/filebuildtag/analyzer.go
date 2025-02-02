// Package filebuildtag exposes the analyzer.Analyzer object and code to use the filebuildtag linter.
package filebuildtag

import (
	"flag"
	"fmt"
	"go/ast"
	"path/filepath"
	"strings"

	"github.com/aziule/filebuildtag/internal"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const (
	// Doc of the linter.
	Doc = `ensure Go files have the expected "// +build <tag>" instruction based on the file name

Bind file names to their expected build tags, such as:
	Files named "foo.go" must have the "foo" build tag
	Files with the suffix "*_integration_test.go" must have the "integration" build tag`
	// FlagFiletagsName is the name of the default filetags flag. It is exported to be reused from linters runners.
	FlagFiletagsName = "filetags"
	// FlagFiletagsDoc is the usage doc of the default filetags flag. It is exported to be reused from linters runners.
	FlagFiletagsDoc = `Comma-separated list of file names and build tags using the form "pattern:tag". For example:
- Single pattern: "*foo.go:tag1"
- Multiple patterns: "*foo.go:tag1,*foo2.go:tag2"`
)

var Analyzer = &analysis.Analyzer{
	Name:     "filebuildtag",
	Doc:      Doc,
	Flags:    flags(),
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func flags() flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.String(FlagFiletagsName, "", FlagFiletagsDoc)
	return *fs
}

func run(pass *analysis.Pass) (interface{}, error) {
	filetags, err := parseFlags(pass.Analyzer.Flags)
	if err != nil {
		return nil, err
	}

	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.File)(nil),
	}
	inspector.Preorder(nodeFilter, func(node ast.Node) {
		f := node.(*ast.File)
		filename := getFilename(pass, f)
		tags := internal.CheckGoFile(pass, f)
		for pattern, tag := range filetags {
			ok, _ := filepath.Match(pattern, filename)
			if !ok {
				continue
			}

			foundTag := false
			for i := range tags {
				if tags[i] == tag {
					foundTag = true
					break
				}
			}

			if !foundTag {
				pass.Reportf(f.Pos(), `missing expected build tag: "%s"`, tag)
			}
		}
	})
	return nil, nil
}

func parseFlags(flags flag.FlagSet) (map[string]string, error) {
	filetags := make(map[string]string)
	f := flags.Lookup(FlagFiletagsName)
	if f == nil {
		return filetags, nil
	}
	args := strings.Split(f.Value.String(), ",")
	for i := 0; i < len(args); i++ {
		filetag := strings.TrimSpace(args[i])
		if filetag == "" {
			continue
		}

		parts := strings.Split(filetag, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf(`malformed argument: "%s", must be of the form "pattern:tag"`, filetag)
		}

		parts[0] = strings.TrimSpace(parts[0])
		parts[1] = strings.TrimSpace(parts[1])
		if parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf(`malformed argument: "%s", must be of the form "pattern:tag"`, filetag)
		}
		filetags[parts[0]] = parts[1]
	}
	return filetags, nil
}

func getFilename(pass *analysis.Pass, file *ast.File) string {
	path := pass.Fset.Position(file.Pos()).Filename
	_, filename := filepath.Split(path)
	return filename
}
