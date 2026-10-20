// titles_csv.go
//
// TEMPORARY: loads already-extracted (and scrubbed) titles from a labels CSV so JEV can be tested
// on a chosen set of rows without re-running the PDF pipeline. Remove once the A/B test is settled.

package main

import (
	"encoding/csv"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var titlesCSVAmount = regexp.MustCompile(`^-?\d+\.\d{2}$`)

// ParseTitlesCSV reads the "date", "amount" and "title" columns by header name; other columns are ignored.
func ParseTitlesCSV(data string) ([]Transaction, error) {
	records, err := csv.NewReader(strings.NewReader(data)).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("titles csv is empty")
	}
	columns := map[string]int{}
	for i, name := range records[0] {
		columns[name] = i
	}
	for _, name := range []string{"date", "amount", "title"} {
		if _, ok := columns[name]; !ok {
			return nil, fmt.Errorf("titles csv has no %q column", name)
		}
	}
	transactions := make([]Transaction, 0, len(records)-1)
	for i, record := range records[1:] {
		amount := strings.TrimSpace(record[columns["amount"]])
		if !titlesCSVAmount.MatchString(amount) {
			return nil, fmt.Errorf("titles csv line %d: bad amount %q", i+2, amount)
		}
		cents, err := strconv.ParseInt(strings.ReplaceAll(amount, ".", ""), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("titles csv line %d: %w", i+2, err)
		}
		transactions = append(transactions, Transaction{
			Date:        record[columns["date"]],
			Description: record[columns["title"]],
			AmountCents: cents,
		})
	}
	return transactions, nil
}
