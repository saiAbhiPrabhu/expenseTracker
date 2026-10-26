# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users
An individual with Bank of America statement PDFs who wants to understand their own spending, privately. They upload statements, check what was extracted and categorized, then read and correct a dashboard of totals.

## Product Purpose
Reads BofA statement PDFs, extracts each transaction, anonymizes it, and has JEV (TypeSafe System One) categorize it and judge whether it is essential. The result feeds a dashboard of categorized spending totals the user can review and edit. Success: the user trusts the numbers and can see where their money goes.

## Positioning
Privacy-first and verifiable together. Name, address and account number are cropped away and descriptions scrubbed locally before anything reaches JEV, and each statement section's printed total must equal the sum of extracted rows. Low-confidence categorizations are flagged for review.

## Operating Context
Flow: Upload, Process, Review, Dashboard (home), Edit data. Input is statement PDFs, read via Poppler and Tesseract OCR. Go backend (`-serve`), React/TypeScript/Vite front end in `web/`. The demo uses only synthetic, committed data and needs no API key.

## Capabilities and Constraints
- Expenditure means any outflow (negative amount). Money-movement categories (Transfers and Peer Payments, Savings and Investing, Debt Payments) can be excluded from dashboard totals.
- 20-category set plus an essential flag; review threshold is confidence < 0.5.
- Not built: real login, real upload (disabled in demo), partial uploads, real statement removal.
- Never use real statements in demo or test data.

## Brand Commitments
`web/DESIGN.md` currently mandates a barebones look (no animation, no new colors or fonts). The user says this is temporary, so it is not a binding brand commitment. Visual direction is open for later work.

## Evidence on Hand
Synthetic demo data only (`demo/`). No testimonials, customers or benchmarks exist; do not fabricate any.

## Product Principles
- Privacy first: sensitive data never leaves the machine unscrubbed.
- Numbers must be verifiable: checksums and review flags over silent guesses.
- The dashboard is home; other steps serve it.
- Heuristics must not be overfit to one person's data.
