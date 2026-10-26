---
name: Expense Tracker
description: A ledger-spread interface in ink on cream for reading, checking and correcting spending totals.
colors:
  cream-page: "#faf6ec"
  ink-navy: "#16203f"
  quiet-navy: "#5a6073"
  red-ink: "#9b1c1c"
typography:
  running-head:
    fontFamily: "Literata, Georgia, serif"
    fontSize: "13px"
    fontWeight: 600
    lineHeight: 1.55
    letterSpacing: "0.2em"
  title:
    fontFamily: "Literata, Georgia, serif"
    fontSize: "30px"
    fontWeight: 600
    lineHeight: 1.2
    letterSpacing: "-0.005em"
  section:
    fontFamily: "Literata, Georgia, serif"
    fontSize: "12.5px"
    fontWeight: 600
    lineHeight: 1.55
    letterSpacing: "0.12em"
  body:
    fontFamily: "Literata, Georgia, serif"
    fontSize: "16px"
    fontWeight: 400
    lineHeight: 1.55
    letterSpacing: "normal"
  table:
    fontFamily: "Literata, Georgia, serif"
    fontSize: "15px"
    fontWeight: 400
    lineHeight: 1.55
    fontFeature: "tnum, lnum"
  label:
    fontFamily: "Literata, Georgia, serif"
    fontSize: "13px"
    fontWeight: 600
    lineHeight: 1.55
    letterSpacing: "0.12em"
rounded:
  none: "0px"
spacing:
  xs: "8px"
  sm: "14px"
  md: "28px"
  lg: "48px"
components:
  button:
    backgroundColor: "{colors.cream-page}"
    textColor: "{colors.ink-navy}"
    typography: "{typography.label}"
    rounded: "{rounded.none}"
    padding: "12px 26px"
  button-disabled:
    backgroundColor: "{colors.cream-page}"
    textColor: "{colors.quiet-navy}"
    rounded: "{rounded.none}"
    padding: "12px 26px"
  field:
    backgroundColor: "{colors.cream-page}"
    textColor: "{colors.ink-navy}"
    rounded: "{rounded.none}"
    padding: "8px 10px"
---

# Design System: Expense Tracker

## Overview

**Creative North Star: "The Ledger Spread"**

The interface is an open book of accounts. The left page is a rail: the five steps and three running counts. The right page is the step at hand. A single ruled line down the middle is the spine. Everything is ink on cream: navy text and rules on a warm off-white page, with nothing filled dark, nothing raised and nothing rounded. Importance is shown by rule weight and by position, never by a second colour or a shadow. The one concession is red ink, kept for what needs attention, as in a real ledger.

The tone is quiet, ruled and countable. Density is that of a working ledger: tables are the content, so type is sturdy enough to read at 15px and figures line up in columns. The login screen is the one expressive moment (a small book that opens in 3D and slides out a bar of links); everything after it is a task tool and moves quickly.

**Key Characteristics:**
- Two inks and a quiet tint: navy on cream, with navy at 70% for secondary text. Red ink only for errors and review counts.
- Line, not fill: hairlines, double rules and a heavier rule carry all hierarchy.
- Square everywhere. No radius, no shadow, no gradient.
- One serif (Literata) in a few weights; spaced capitals only for short labels.
- Motion is short (200 to 250ms). Only the book and its bar take about 600ms.

## Colors

Two inks on one paper, plus a tint and a signal.

### Primary
- **Ink Navy** (#16203f): all text, rules, outlines, the book icon and the filled share of every bar. The only dark colour.

### Secondary
- **Red Ink** (#9b1c1c): error labels and the left rule on error notices, and the counts of rows that need review (in the rail and on the month tabs). 7.5:1 on cream.

### Neutral
- **Cream Page** (#faf6ec): the page, every control, every table. Nothing sits on a darker surface.
- **Quiet Navy** (#5a6073): navy at 70% over cream, 5.8:1. Captions, notices, disabled reasons, dates on phones. Faint navy (50% opacity) is used only for things that are genuinely disabled or not yet reached.
- **Rule Tint** (navy at 26% over cream): the hairlines between table rows and under section headings.

### Named Rules
**The Two Inks Rule.** Navy carries everything; red ink is the only other colour and always travels with a text label or a rule, so colour is never the sole signal.

**The Weight-Not-Colour Rule.** If something matters more, make its rule heavier, double or longer. Do not tint it.

## Typography

**Display Font:** Literata (variable, weight and optical size), with Georgia as fallback
**Body Font:** Literata, same file
**Label Font:** Literata, in spaced capitals at medium weight

**Character:** A sturdy, screen-designed serif with a large x-height. It keeps its strokes at 12 to 15px, which the earlier high-contrast Didone could not.

### Hierarchy
- **Running head** (600, 13px, 0.2em tracking, caps): the "Demo" line above the page rule.
- **Title** (600, 30px, 1.2): the step title, in sentence case ("Step 3 of 5: Review").
- **Section** (600, 12.5px, 0.12em, caps): table and section headings, with a hairline beneath.
- **Body** (400, 16px, 1.55): prose and form labels. Step instruction line is 17px in quiet navy.
- **Table** (400, 15px, tabular lining figures): all table cells; numbers right-aligned.
- **Label** (600, 13px, 0.12em, caps): buttons, tabs, "Show…" bars, table headers (12px, 0.1em).

### Named Rules
**The Short-Caps Rule.** Spaced capitals are for words of a few letters. Category names, merchants and descriptions stay in sentence case.

**The Tabular Rule.** Every figure column uses tabular lining numerals, right-aligned.

## Layout

A two-page spread on screens wider than 900px: a 252px rail on the left, the step page on the right, with the page's content padded 48px from the spine rule. The page is up to 1240px wide. Below 900px the rail becomes a wrapped row of step links above the content and the running counts are hidden.

The step page has a title, a one-line instruction, the step content, and a sticky action bar pinned to the bottom (Back on the left, the primary action beside it, a disabled reason beside that). The bar sits under a single rule and outside any frame.

The Dashboard is a two-column grid above 1100px (category ledger about 1.55fr, month and necessity about 1fr, 56px gap) and stacks below. Review and Edit tables stack each row into a small block at 520px and below. Tables never scroll sideways.

### Named Rules
**The One Primary Place Rule.** The primary action is the right-hand button in the action bar. It is not louder than Back; its place is its priority.

## Elevation & Depth

Flat. There are no shadows and no layering. Depth is a matter of rules: the only exception is an inset left bar on a changed row, drawn with an inset box-shadow of 4px solid navy (it is a line, not a shadow). Sticky elements (the rail, the filters on Edit, the action bar) sit on the cream page with a rule on the edge they touch.

### Named Rules
**The No-Raise Rule.** Nothing is lifted, glowing or blurred. If something must stand apart, it gets a rule.

## Shapes

Square corners everywhere (0px). Three line weights carry the form language:
- **Double outline:** a 1px border plus a second 1px outline 3px outside it. Used for buttons and the drawer menu bar.
- **Single hairline:** a 1px navy border. Used for fields, selects, checkboxes, "Show…" sections, the dropzone and image frames.
- **Heavy rule:** a double 3px rule under the current step, tab and table header, and a 2 to 4px bar for errors and changed rows.

### Named Rules
**The Frame Means Action Rule.** Only things you press get the double outline. Containers use hairlines.

## Components

### Buttons
- **Shape:** square, 1px navy border plus a 1px outline 3px outside (double outline).
- **Style:** cream fill, navy text, 13px 600 spaced capitals, padding 12px 26px. One style for all buttons; small ones (Undo) drop to 11.5px.
- **Hover:** the outer outline drifts out to 5px over 220ms.
- **Focus:** a 3px solid outline at 5px offset, deliberately different from hover.
- **Disabled:** 50% opacity with a dashed outer outline; a quiet reason sits beside it.

### Tabs
Months in Review and Pages / Text from pages / Rows in Process. Spaced-caps text over a single navy rule, no box. The open tab has a double rule beneath. Review tabs carry a count in red ink.

### Fields
Cream, 1px navy border, no radius. Selects show a thin navy chevron. Checkboxes are 20px hairline squares, holding a navy tick when checked. The file input's button follows the same hairline.

### Tables
No fills or stripes. A double rule under the header, hairlines between rows, a double rule above totals. Numbers right-aligned in tabular figures. A changed row has a 4px navy bar on its left edge.

### "Show…" sections
A hairline bar with the title in spaced capitals and a plus that becomes a minus when open (two drawn lines, not a glyph). Content opens downward over 220ms and the opened content sits under a hairline.

### Bars
A hairline navy track across the full extent with the share drawn as a thicker (4px) navy rule. Numbers always sit beside the bar. The essential / non-essential split uses the same device.

### Rail and action bar
The rail lists the five steps (current one bold with a double rule under it, reached steps as faint-underlined links, unreached steps at half strength), then three counts: statements, rows to review (red when above zero) and rows changed. The action bar is sticky at the bottom of the page.

### Book menu (signature component)
A small navy book at the top left. Clicking it opens the cover in 3D, flips three pages, and drifts the book sideways while a double-outlined bar of links slides out and fades in. About 600ms, ease-in-out. It closes in reverse. It is a real button named "Menu" with `aria-expanded`, and Esc closes it.

## Do's and Don'ts

### Do:
- **Do** show importance with rule weight, position and type weight.
- **Do** keep every figure right-aligned in tabular lining numerals.
- **Do** use the double outline only on things you press, hairlines on containers.
- **Do** give red ink a text label or a rule alongside it.
- **Do** keep state-change motion to 200 to 250ms and honour `prefers-reduced-motion` (everything becomes instant).

### Don't:
- **Don't** use fills, shadows, gradients or rounded corners.
- **Don't** add colours: navy, its tints and red ink are the whole palette.
- **Don't** set long labels, categories or descriptions in spaced capitals.
- **Don't** nest a double outline inside another double outline.
- **Don't** loop any animation or use a spinner; loading is a static notice.
- **Don't** let a table scroll sideways; stack its rows instead.
