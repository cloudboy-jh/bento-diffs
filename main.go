package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudboy-jh/bento-diffs/internal/adapter"
	"github.com/cloudboy-jh/bento-diffs/internal/model"
	"github.com/cloudboy-jh/bento-diffs/internal/parser"
	"github.com/cloudboy-jh/bentotui/theme"
)

type options struct {
	layout      string
	themeName   string
	noSyntax    bool
	context     int
	noLineNums  bool
	mock        bool
	patchPath   string
	beforePath  string
	afterPath   string
	stdinSource bool
}

func main() {
	opt, err := parseFlags()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}

	if _, err := theme.SetTheme(opt.themeName); err != nil {
		fmt.Fprintf(os.Stderr, "error: unknown theme %q\n", opt.themeName)
		fmt.Fprintf(os.Stderr, "available themes: %s\n", strings.Join(theme.AvailableThemes(), ", "))
		os.Exit(2)
	}

	var diffs []parser.DiffResult
	if opt.mock {
		diffs, err = mockDiffs(opt.context)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: mock diff generation failed:", err)
			os.Exit(1)
		}
	} else {
		patch, fileName, loadErr := loadPatch(opt)
		if loadErr != nil {
			fmt.Fprintln(os.Stderr, "error:", loadErr)
			os.Exit(1)
		}

		diffs, err = parser.ParseUnifiedDiffs(patch)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: parse diff:", err)
			os.Exit(1)
		}
		if len(diffs) == 0 {
			fmt.Fprintln(os.Stderr, "error: no file diffs found")
			os.Exit(1)
		}

		if len(diffs) == 1 && diffs[0].DisplayFile == "" && fileName != "" {
			diffs[0].DisplayFile = fileName
		}
	}

	layout := model.LayoutSplit
	if opt.layout == "stacked" {
		layout = model.LayoutStacked
	}

	workspace := adapter.NewWorkspaceAdapter(diffs)
	app := model.New(workspace, theme.CurrentTheme(), model.Options{
		Layout:          layout,
		SyntaxEnabled:   !opt.noSyntax,
		ShowLineNumbers: !opt.noLineNums,
		UseTTYInput:     opt.stdinSource,
	})

	if err := app.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func parseFlags() (options, error) {
	var opt options
	flag.StringVar(&opt.layout, "layout", "split", "initial layout: split|stacked")
	flag.StringVar(&opt.themeName, "theme", "catppuccin-mocha", "theme preset name")
	flag.BoolVar(&opt.noSyntax, "no-syntax", false, "disable syntax highlighting")
	flag.IntVar(&opt.context, "context", 3, "context lines around hunks")
	flag.BoolVar(&opt.noLineNums, "no-line-numbers", false, "hide line number column")
	flag.BoolVar(&opt.mock, "mock", false, "run built-in mock diff in TUI")
	flag.StringVar(&opt.patchPath, "patch", "", "read unified diff from patch file")
	flag.Parse()

	if opt.layout != "split" && opt.layout != "stacked" {
		return opt, errors.New("--layout must be split or stacked")
	}

	args := flag.Args()
	if len(args) == 2 {
		opt.beforePath = args[0]
		opt.afterPath = args[1]
	} else if len(args) != 0 {
		return opt, errors.New("expected two positional file arguments or none")
	}

	st, err := os.Stdin.Stat()
	if err == nil && (st.Mode()&os.ModeCharDevice) == 0 {
		opt.stdinSource = true
	}

	if !opt.mock && opt.patchPath == "" && opt.beforePath == "" && !opt.stdinSource {
		return opt, errors.New("no input provided (pipe diff, use --patch, or pass two files)")
	}

	if opt.patchPath != "" && opt.beforePath != "" {
		return opt, errors.New("--patch cannot be combined with two-file mode")
	}

	return opt, nil
}

func loadPatch(opt options) (patch string, fileName string, err error) {
	if opt.patchPath != "" {
		b, readErr := os.ReadFile(opt.patchPath)
		if readErr != nil {
			return "", "", readErr
		}
		return string(b), "", nil
	}

	if opt.beforePath != "" {
		before, readErr := os.ReadFile(opt.beforePath)
		if readErr != nil {
			return "", "", readErr
		}
		after, readErr := os.ReadFile(opt.afterPath)
		if readErr != nil {
			return "", "", readErr
		}
		name := filepath.Base(opt.afterPath)
		patch, _, _ := parser.GenerateDiff(string(before), string(after), name, opt.context)
		return patch, name, nil
	}

	b, readErr := io.ReadAll(os.Stdin)
	if readErr != nil {
		return "", "", readErr
	}
	return string(b), "", nil
}

func mockDiffs(context int) ([]parser.DiffResult, error) {
	beforeGo := strings.Join([]string{
		"package main",
		"",
		"import \"fmt\"",
		"",
		"func main() {",
		"    fmt.Println(\"hello\")",
		"}",
	}, "\n") + "\n"

	afterGo := strings.Join([]string{
		"package main",
		"",
		"import (",
		"    \"fmt\"",
		"    \"time\"",
		")",
		"",
		"func main() {",
		"    fmt.Println(\"hello from bento-diffs\")",
		"    fmt.Println(time.Now().Format(time.RFC3339))",
		"}",
	}, "\n") + "\n"

	beforeReadme := strings.Join([]string{
		"# Demo",
		"",
		"A tiny demo file.",
		"",
		"- first item",
	}, "\n") + "\n"

	afterReadme := strings.Join([]string{
		"# Demo",
		"",
		"A tiny demo file for bento-diffs.",
		"",
		"- first item",
		"- second item",
	}, "\n") + "\n"

	p1, _, _ := parser.GenerateDiff(beforeGo, afterGo, "cmd/demo/main.go", context)
	p2, _, _ := parser.GenerateDiff(beforeReadme, afterReadme, "README.md", context)
	patch := p1 + "\n" + p2
	return parser.ParseUnifiedDiffs(patch)
}
