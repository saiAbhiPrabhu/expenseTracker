# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Prospective end users: people who might upload their own bank statements to understand where their money goes. They meet the product through a demo that runs on synthetic data, so they are judging whether they would trust it with a real statement.

## Product Purpose

Reads a Bank of America statement PDF, extracts each transaction, anonymizes it, and has JEV (TypeSafe's System One model) categorize it and judge whether it is an essential expense, feeding a dashboard. The demo walks a visitor through the whole path. Success is equally two things: the pipeline is transparent enough to trust, and the dashboard leaves the visitor knowing where the money goes.

## Positioning

Every stage between the PDF and the dashboard can be inspected: the cropped page, the OCR text, the extracted rows with per-section checksums, the model's categories with low-confidence rows flagged for review, and editable data. The name and account number are cropped out before any text is produced, and descriptions are scrubbed before JEV sees them.

## Operating Context

The visitor moves through Upload, Process, Review, Dashboard and Edit data. The demo replays a saved JEV run on four fictional statements; there are no live model calls and no per-user server state.

## Capabilities and Constraints

- Fixed five-step flow: Upload, Process, Review, Dashboard (home after the first pass), Edit data ("Done" returns to the dashboard). This order is a confirmed constraint.
- "Expenditure" means any outflow (negative amount). Money-movement categories (Transfers and Peer Payments, Savings and Investing, Debt Payments) can be left out of dashboard totals.
- Review flag when JEV confidence is below 0.5; essential when the probability is 0.5 or more.
- The log-in page is intentionally non-functional; there are no accounts.
- Stack: Go server with a React, TypeScript and Vite front end in `web/`.
- Not confirmed as binding: platform limits beyond web, and whether privacy framing must stay central in the UI.

## Evidence on Hand

Synthetic data only, committed in `demo/`: 200 transactions over four months, four fictional statement PDFs, saved page images and OCR artifacts, and a saved JEV run. There are no real customers, testimonials or benchmarks; none may be invented.

## Product Principles

- Show the work: each stage of reading a statement is visible and checkable.
- The dashboard answers "where does my money go" without explanation.
- One primary action per step, in a fixed and obvious order.
- Few words: labels and numbers over sentences.

## Accessibility & Inclusion

No product-specific standard established. The existing rules already require named controls, labelled tables and real buttons.
