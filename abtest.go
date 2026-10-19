// abtest.go
//
// Responsibility: compare the 20-category and 12-category answers for the same transactions and
// render the comparison. Pure functions only; running the calls and writing files lives in main.go.

package main

import (
	"fmt"
	"strings"
)

// categoryIn12 maps each 20-set category to its 12-set bucket, so the two sets can be compared.
// "Fees and Insurance" has no clear 12-set bucket, so it is absent and reported as unmapped.
var categoryIn12 = map[string]string{
	"Housing":                     "Housing and Utilities",
	"Utilities and Internet":      "Housing and Utilities",
	"Groceries":                   "Groceries",
	"Dining and Drinks":           "Dining and Drinks",
	"Transportation":              "Transportation",
	"Health and Pharmacy":         "Health and Wellness",
	"Fitness and Wellness":        "Health and Wellness",
	"Education":                   "Education",
	"Clothing and Accessories":    "Shopping",
	"General Shopping":            "Shopping",
	"Entertainment":               "Entertainment and Subscriptions",
	"Subscriptions":               "Entertainment and Subscriptions",
	"Personal Care":               "Personal Care, Travel, and Gifts",
	"Travel":                      "Personal Care, Travel, and Gifts",
	"Gifts and Donations":         "Personal Care, Travel, and Gifts",
	"Savings and Investing":       "Savings, Investing, and Debt",
	"Debt Payments":               "Savings, Investing, and Debt",
	"Transfers and Peer Payments": "Transfers and Peer Payments",
	"Income and Refunds":          "Income and Refunds",
	catchAllCategory:              catchAllCategory,
}

type ABReport struct {
	Total           int
	Unmapped        int // 20-set answer has no 12-set equivalent
	Agreements      int // among mapped rows, 20-set answer maps to the 12-set answer
	AvgConfidence20 float64
	AvgConfidence12 float64
	Rows            []ABRow
}

type ABRow struct {
	Result     ClassifiedTransaction
	MappedTo12 string // "" when unmapped
	Agrees     bool
}

func CompareCategorySets(results []ClassifiedTransaction) ABReport {
	report := ABReport{Total: len(results)}
	for _, result := range results {
		row := ABRow{Result: result, MappedTo12: categoryIn12[result.Category20.Choice]}
		switch {
		case row.MappedTo12 == "":
			report.Unmapped++
		case row.MappedTo12 == result.Category12.Choice:
			row.Agrees = true
			report.Agreements++
		}
		report.AvgConfidence20 += result.Category20.Confidence
		report.AvgConfidence12 += result.Category12.Confidence
		report.Rows = append(report.Rows, row)
	}
	if report.Total > 0 {
		report.AvgConfidence20 /= float64(report.Total)
		report.AvgConfidence12 /= float64(report.Total)
	}
	return report
}

// FormatABReport renders a summary plus one table row per transaction for manual review.
func FormatABReport(report ABReport) string {
	var out strings.Builder
	mapped := report.Total - report.Unmapped
	fmt.Fprintf(&out, "# A/B: categories_20 vs categories_12\n\n")
	fmt.Fprintf(&out, "- Transactions: %d\n", report.Total)
	fmt.Fprintf(&out, "- Agreement (20-set mapped to 12-set): %d of %d comparable (%d unmapped)\n", report.Agreements, mapped, report.Unmapped)
	fmt.Fprintf(&out, "- Mean confidence: 20-set %.2f, 12-set %.2f\n\n", report.AvgConfidence20, report.AvgConfidence12)
	fmt.Fprintf(&out, "| Date | Amount | Description | 20-set (conf) | 12-set (conf) | Agree | P(essential) |\n|---|---|---|---|---|---|---|\n")
	for _, row := range report.Rows {
		transaction, result := row.Result.Transaction, row.Result
		fmt.Fprintf(&out, "| %s | %s | %s | %s (%.2f) | %s (%.2f) | %s | %.2f |\n",
			transaction.Date, formatDollars(transaction.AmountCents), escapeTableCell(transaction.Description),
			result.Category20.Choice, result.Category20.Confidence,
			result.Category12.Choice, result.Category12.Confidence,
			agreementMark(row), result.Essential.Noul)
	}
	return out.String()
}

func agreementMark(row ABRow) string {
	switch {
	case row.MappedTo12 == "":
		return "n/a"
	case row.Agrees:
		return "yes"
	}
	return "NO"
}

func escapeTableCell(text string) string {
	return strings.ReplaceAll(text, "|", "\\|")
}
