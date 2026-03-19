package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	bentodiffs "github.com/cloudboy-jh/bento-diffs"
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

	runOpts := bentodiffs.DefaultOptions()
	runOpts.Layout = opt.layout
	runOpts.ThemeName = opt.themeName
	runOpts.SyntaxEnabled = !opt.noSyntax
	runOpts.ShowLineNumbers = !opt.noLineNums
	runOpts.UseTTYInput = opt.stdinSource

	if opt.mock {
		diffs, err := bentodiffs.MockDiffs(opt.context)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: mock diff generation failed:", err)
			os.Exit(1)
		}
		if err := bentodiffs.RunDiffs(diffs, runOpts); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return
	}

	patch, fileName, err := loadPatch(opt)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if err := bentodiffs.RunPatch(patch, fileName, runOpts); err != nil {
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
		patch, _, _ := bentodiffs.GenerateDiff(string(before), string(after), name, opt.context)
		return patch, name, nil
	}

	b, readErr := io.ReadAll(os.Stdin)
	if readErr != nil {
		return "", "", readErr
	}
	return string(b), "", nil
}
