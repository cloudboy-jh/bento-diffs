# Architecture split with BentoTUI

This repository stays a standalone tool and keeps all diff intelligence local.

## Boundary contract

`bentodiffs` owns:
- parsing unified diffs (`pkg/bentodiffs/parser`)
- hunk and line pairing logic
- intraline change detection
- syntax highlighting + diff line rendering (`pkg/bentodiffs/renderer`)
- workspace DTO adapter that prepares render-ready UI payloads (`pkg/bentodiffs/workspace`)

`bentotui` owns:
- Bricks (official reusable components)
- Recipes (copy/own composed interaction flows)
- Rooms (named layout contracts)
- Bento app templates and theme contracts

`bentodiffs` consumes BentoTUI as a platform dependency only. No parser/syntax/intraline behavior is moved into BentoTUI.

## DTO interface used by UI

The UI model consumes `workspace.WorkspaceDTO` from `pkg/bentodiffs/workspace`, which maps directly to the Bento diff workspace pattern:

- `HeaderDTO` -> header row (file + stats + layout)
- `FileRailDTO` -> file index entries (used for navigation state)
- `MainDiffPaneDTO` -> render-ready diff lines for active file
- `FooterStatusDTO` -> status/command cards

The adapter renders the active file to lines using parser + renderer logic, then UI bricks only handle layout, sizing, and interaction state.

## How to consume with Bento v0.5.3

1. Build a workspace adapter from parsed diffs:
   - `workspace.New(diffs)`
2. Ask for workspace DTOs as app state changes:
   - `Build(activeFile, workspace.RenderOptions{...})`
3. Compose UI with Bento-first contracts:
	- `rooms.Focus(...)` for the rail-less diff app shell
	- Recipes-first for app flows:
	  - `recipes/filterbar` for filter/search interaction
	  - `recipes/emptystatepane` for no-data and no-match states
	  - `recipes/commandpaletteflow` for command launcher/open routing
	- Bricks for low-level rendering (`bar`, `list`, `dialog`, etc.)

## Layering policy in bentodiffs

- Bricks: low-level, visual units (file header, diff pane, optional file list)
- Recipes: copy/own flow composition and main extension surface
- Rooms: explicit layout contracts per page (`rooms.Focus` for app shell)

When adding a new UI flow, prefer this order:
1. existing official recipe
2. local custom recipe in `recipes/`
3. existing brick composition
4. brand new low-level component only when necessary

This keeps product iteration in `bentodiffs` without changing BentoTUI internals.

Current app default is rail-less: the persistent left rail is hidden and file navigation is routed through recipes (`command-palette-flow` and `filter-bar`).

This keeps composition Bento-first while preserving domain logic ownership in `bentodiffs`.
