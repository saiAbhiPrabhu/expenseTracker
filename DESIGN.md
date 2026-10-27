---
name: Expense Tracker
description: A Dieter Rams instrument for reading bank statements: pale neutral surfaces, near-black type, one teal-green signal that only marks controls.
colors:
  ground: "#e4e4e0"
  surface: "#f5f5f3"
  field: "#ffffff"
  ink: "#1a1a19"
  ink-secondary: "#53534f"
  rule: "#c8c8c3"
  track: "#d6d6d1"
  changed: "#e8e8e3"
  signal: "#00786a"
  signal-press: "#005a4f"
  on-signal: "#ffffff"
typography:
  title:
    fontFamily: "Helvetica Neue, Helvetica, Arial, system-ui, sans-serif"
    fontSize: "1.75rem"
    fontWeight: 700
    lineHeight: 1.2
  heading:
    fontFamily: "Helvetica Neue, Helvetica, Arial, system-ui, sans-serif"
    fontSize: "1.25rem"
    fontWeight: 700
    lineHeight: 1.3
  body:
    fontFamily: "Helvetica Neue, Helvetica, Arial, system-ui, sans-serif"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.5
  label:
    fontFamily: "Helvetica Neue, Helvetica, Arial, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 700
    lineHeight: 1.5
  caption:
    fontFamily: "Helvetica Neue, Helvetica, Arial, system-ui, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 400
    lineHeight: 1.5
rounded:
  key: "2px"
spacing:
  xs: "4px"
  sm: "8px"
  md: "16px"
  lg: "24px"
  xl: "32px"
  xxl: "48px"
components:
  key:
    backgroundColor: "{colors.field}"
    textColor: "{colors.ink}"
    rounded: "{rounded.key}"
    height: "40px"
    padding: "8px 16px"
  key-primary:
    backgroundColor: "{colors.signal}"
    textColor: "{colors.on-signal}"
    rounded: "{rounded.key}"
    height: "40px"
    padding: "8px 16px"
  key-primary-press:
    backgroundColor: "{colors.signal-press}"
  step-key-current:
    backgroundColor: "{colors.signal}"
    textColor: "{colors.on-signal}"
    rounded: "{rounded.key}"
    height: "40px"
  surface-panel:
    backgroundColor: "{colors.surface}"
    rounded: "{rounded.key}"
    padding: "32px"
---

# Design System: Expense Tracker

## Overview

**Creative north star: the instrument.** The front end follows Dieter Rams's "less, but better". It is a tool, not a page: a fixed control column holds the five step keys and the current step's Back and primary buttons, and a flat working surface beside it holds the step's content. The mechanism is always on view and the primary action never moves.

The product is judged by whether someone would trust it with a real statement, so the interface stays quiet and every stage of reading a statement shows as a plain, checkable table. Nothing decorates, nothing moves.

**Key characteristics:**
- Pale "glacial" neutrals, near-black type, flat 1px rules.
- One signal color, on controls only.
- A 4px spacing base, a five-step type scale, two font weights.
- No shadows, gradients, icons, illustration, animation or webfonts.

## Colors

A neutral palette with a single bold teal-green. Color is a signal for controls, never decoration.

- **Ground** (`#e4e4e0`): the page behind everything.
- **Surface** (`#f5f5f3`): the control column and the working surface.
- **Field** (`#ffffff`): keys, selects and inputs, so controls read as parts sitting on the surface.
- **Ink** (`#1a1a19`) and **Ink secondary** (`#53534f`): all text; secondary is for instructions, reasons and captions (7:1 on the surface).
- **Rule** (`#c8c8c3`) and **Track** (`#d6d6d1`): borders, row rules, and the light track behind bars.
- **Changed** (`#e8e8e3`): the tint of an edited row, together with bold text.
- **Signal** (`#00786a`, pressed `#005a4f`, text on it `#ffffff`, about 5.4:1): the primary action, the current step key, keyboard focus, checkbox accent, text selection.

**The Signal rule.** Teal-green marks something you can act on or where you are. It never fills a bar, a heading or a flag. Review and edit states are carried by tint and weight, not color.

## Typography

One family, the system sans stack (`Helvetica Neue`, Helvetica, Arial, system-ui), so nothing is downloaded. Two weights, 400 and 700.

Scale: title 1.75rem, heading 1.25rem, body 1rem, label 0.875rem, caption 0.75rem. Table figures use tabular numerals and sit right-aligned, with their column headers aligned to them. Body text is capped at 70ch. OCR text is the one place monospace is used, because it is data.

## Layout

A 4px base with named steps: 4, 8, 16, 24, 32, 48. Everything aligns to those.

- **Desktop:** a 240px control column on the left, sticky, and the working surface filling the rest, inside a 1280px page. The top bar holds the wordmark and the three site links, always visible.
- **Control column order:** step keys, a rule, Back and the primary key, the reason a primary key is disabled directly under it, then "Sample data only."
- **Narrow screens (820px and below):** the step keys become a five-across strip, and the Back and primary keys fix to the bottom edge of the screen. Review and edit rows reflow into short blocks; nothing scrolls the page sideways.
- Home and Log in use a single 480px surface aligned to the same grid.

## Elevation & Depth

None. Hierarchy comes from 1px rules, surface tone and the edge between a key and its panel. Pressed keys change tone. Shadows are not used.

## Shapes

Plain rectangles with a 2px corner. Bars are flat rectangles, 8px high: near-black fill on the light track.

## Components

- **Key** (button or link-button): white field, 1px ink border, 700 weight, 40px minimum height. Hover and press shift the tone. Disabled keys lose their fill and border contrast and the reason is written next to them.
- **Primary key**: the Signal fill. Only one per screen, always in the control column; Home has one for "Start demo".
- **Step key**: numbered, in the control column. The current step is Signal-filled; a reached step is a key; a step not yet reached is flat grey text. Not-yet-reached steps are not clickable.
- **Table**: header rule in ink, row rules in Rule, right-aligned tabular numbers, a bold footer for totals.
- **Bars**: flat ink rectangles on Track, always beside the numbers they show.
- **Disclosure** (`details`): 1px bordered, summary in bold; used for secondary detail that is not needed every time.
- **Fields**: 36px, 1px secondary-ink border, 2px corner; disabled fields lose their fill.

## Do's and Don'ts

**Do**
- Keep one primary action per screen, in the control column, and say why it is disabled.
- Use Signal only for controls and position.
- Label actions by what they do; keep text short.
- Keep focus visible: a 2px Signal outline with a 2px offset.

**Don't**
- Don't add shadows, gradients, icons, glass or animation.
- Don't add a second accent color or color a bar.
- Don't hide navigation behind a menu button.
- Don't put a second primary key on a screen.
