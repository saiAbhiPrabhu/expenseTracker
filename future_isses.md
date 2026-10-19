# Future Issues

Known problems and intended fixes. Nothing here is implemented yet.

## Incomplete OCR

### The problem

The July statement fails: one Zelle row (07/16) printed `-177.41` was OCR'd as `-177.4)`. A date-led line without a valid amount is rejected on purpose (skipping it would lose a transaction and break the totals), so one bad character blocks all 79 rows. The true value was recovered by hand: only `-177.41` made the section sum match the printed total.

### Allow partial uploads

One OCR failure should not cancel the upload. Keep every row that parsed, set aside the failed rows with their position, and report how far off the section checksum is.

### Prompt for user intervention

A failed row should ask the user to step in, not end the run:
- Show the row number and reason, using a masked line (today's error prints the raw row).
- Let the user enter the correct amount or confirm a checksum-derived suggestion, and record that it was repaired.
- Re-run the section checksum, then continue.

---
