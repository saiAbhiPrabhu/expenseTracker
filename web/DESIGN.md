# Design language and contributor rules

The demo page is barebones on purpose. **Do not make design decisions.** Everything new must look like what is already here, and nothing animates.

## The design language (as it is today, in `src/styles.css`)
- System font, black text (`#222`) on white. Secondary text is grey (`#666`, class `.notice`).
- One centered column, max width 960px.
- Borders are 1px solid (`#ccc` around panels, `#e3e3e3` between table rows, `#222` under table headers); corners are 4px.
- Buttons are plain outlined buttons (class `.button` for links, plain `<button>` otherwise). Disabled buttons are 50% opaque.
- Tables are plain: left-aligned text, numbers right-aligned (`.amount`), no zebra stripes, no color. Use `.data-table`.
- The only emphasis colors are `#222` (strong) and `#f6f6f6` (a changed or selected row, with a 3px `#222` bar on its left, class `.edited`; drawn with an inset box-shadow, the only shadow allowed).
- Graphs are plain bars made from HTML (a `div` whose width is the share, filled `#222` on a `#e3e3e3` track). No chart library, no SVG effects, no gradients, no shadows, no rounded fills.

## Hard rules
1. **No animations or transitions.** Do not write `transition`, `animation`, `@keyframes`, `transform` effects, smooth scrolling or hover effects. `<details>` is allowed (it does not animate).
2. **No new colors, fonts, icons, images (other than the statement page images the Process step shows), shadows or font sizes.** If you think something needs one, leave it out and mention it in your report.
3. **No new npm dependencies** unless there is no other way. If you add one, say so loudly in your report.
4. **Reuse components:** `ControlBar` (the Back / Continue bar) and the classes above. Do not invent a second way to navigate.
5. **Comfort.** A feature must (a) sit in the page flow with the same table style, (b) say what it is without help from prose, and (c) if it is not useful every time, be hidden by default inside a `<details>` element with a clear summary line ("Show category by month").
6. **Few words.** Cut any text that repeats its heading, the step instruction or the control it sits beside. Keep a caption only for a definition or a scale ("Bars are relative to the largest month"). Prefer "Label: value" lines to sentences. Put the number next to the thing it describes (the money-movement amount sits inside its checkbox label).
7. **Order must be obvious.** Every step has one primary action, always in the `ControlBar`. If it is disabled, say why next to it (`disabledReason`).
8. Keyboard and screen readers: every control needs a visible or `aria-label` name; tables need an `aria-label`; use real `<button>`s.

## The flow
`Upload -> Process -> Review -> Dashboard -> Edit data`. The first time through, the user moves forward with the primary button. After that, **the dashboard is home**: it has an "Edit data" button, and "Done" on the Edit page always returns to the dashboard. Completed steps are clickable in the step list. See `src/flow/`.

## File ownership (so parallel work never collides)
Each step lives in its own folder and is owned by one contributor at a time:
`src/steps/upload/`, `src/steps/process/` (they share the added statements through `src/demo/StatementsContext.tsx`), `src/steps/review/`, `src/steps/dashboard/`, `src/steps/edit/`.
- Put a step's own CSS in a file inside its folder (for example `src/steps/dashboard/dashboard.css`) and import it from the step. Do not edit `src/styles.css` or `src/flow/flow.css`.
- Put a step's own API calls in a file inside its folder (for example `src/steps/dashboard/api.ts`). Do not edit `src/api.ts`.
- Do not edit `src/flow/*`, `src/demo/DemoContext.tsx`, `src/types.ts` or `src/format.ts`. If you need a change there, list it in your report.
- Go side: each step may own one `demo_<step>.go` file plus its test. Do not edit `demo.go`, `main.go` or `serve.go`.

## Checks before you finish
`cd web && npm run build` (this type-checks and builds) and `go vet ./... && go test ./...` from the repo root. Test in a real browser (Chrome headless with puppeteer-core, see the briefing for the path).
