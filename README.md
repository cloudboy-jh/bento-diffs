# bento-diffs

Simple terminal diff viewer that works as both a CLI app and a Go library.

## Repo layout

```text
cmd/bento-diffs/   CLI entrypoint
internal/          parser, renderer, adapter, and app internals
bricks/            local UI components
recipes/           local interaction flows
docs/              architecture notes
```

## CLI

Run the mock view:

```bash
go run ./cmd/bento-diffs --mock
```

Build and run:

```bash
go build -o bento-diffs ./cmd/bento-diffs
./bento-diffs --mock
```

Run with two files:

```bash
go run ./cmd/bento-diffs before.txt after.txt
```

Run with a patch file:

```bash
go run ./cmd/bento-diffs --patch changes.diff
```

## Library

```go
package main

import (
	"log"

	bentodiffs "github.com/cloudboy-jh/bento-diffs"
)

func main() {
	opts := bentodiffs.DefaultOptions()
	diffs, err := bentodiffs.ParseUnifiedDiffs("diff --git a/a.txt b/a.txt\n--- a/a.txt\n+++ b/a.txt\n@@ -1 +1 @@\n-old\n+new\n")
	if err != nil {
		log.Fatal(err)
	}
	if err := bentodiffs.RunDiffs(diffs, opts); err != nil {
		log.Fatal(err)
	}
}
```
