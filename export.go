// export.go
//
// Responsibility: everything the app can export: a stable key per transaction, the raw JEV answers for
// every transaction, and the expenditure totals. Pure functions that return text; main.go writes the files.

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"
	"time"
)

// exportedState is one transaction with JEV's raw 20-category and essential answers.
type exportedState struct {
	Key        string       `json:"key"`
	Date       string       `json:"date"`
	Title      string       `json:"title"`
	Amount     json.Number  `json:"amount"`
	Category20 ChoiceAnswer `json:"category_20"`
	Essential  NoulAnswer   `json:"essential"`
}

type totalsExport struct {
	ExpenditureByMonth    map[string]json.Number `json:"expenditure_by_month"`
	ExpenditureByCategory map[string]json.Number `json:"expenditure_by_category"`
	TotalExpenditure      json.Number            `json:"total_expenditure"`
}

// ExportFiles is the content of every file the export writes.
type ExportFiles struct {
	StatesJSON string
	StatesCSV  string
	TotalsJSON string
}

// Function Contract:
// Input: classified transactions
// Output: the content of the three export files
// Logic: Computer lays out the raw JEV answers for every transaction (JSON and CSV) and the
// expenditure totals computed from the final decisions.
func Export(classified []ClassifiedTransaction) (ExportFiles, error) {
	statesJSON, statesCSV, err := ExportStates(classified)
	if err != nil {
		return ExportFiles{}, err
	}
	decisions := make([]Decision, 0, len(classified))
	for _, c := range classified {
		decision, err := Decide(c)
		if err != nil {
			return ExportFiles{}, err
		}
		decisions = append(decisions, decision)
	}
	totals, err := ComputeTotals(decisions)
	if err != nil {
		return ExportFiles{}, err
	}
	totalsJSON, err := ExportTotals(totals)
	if err != nil {
		return ExportFiles{}, err
	}
	return ExportFiles{StatesJSON: statesJSON, StatesCSV: statesCSV, TotalsJSON: totalsJSON}, nil
}

// Function Contract:
// Input: classified transactions
// Output: every transaction, oldest first, with its key and JEV's raw answers, as JSON and as CSV
// Logic: Computer orders the transactions by date (ties keep their input order) and writes the full
// answers (choice, confidence, all probabilities, essential probability) as JSON, and the headline
// values (category, its confidence, essential probability) as CSV.
func ExportStates(classified []ClassifiedTransaction) (statesJSON string, statesCSV string, err error) {
	ordered, err := sortedByDate(classified)
	if err != nil {
		return "", "", err
	}
	states := make([]exportedState, 0, len(ordered))
	for _, c := range ordered {
		key, err := TransactionKey(c.Transaction)
		if err != nil {
			return "", "", err
		}
		states = append(states, exportedState{
			Key: key, Date: c.Transaction.Date, Title: c.Transaction.Description,
			Amount: formatDollars(c.Transaction.AmountCents), Category20: c.Category20, Essential: c.Essential,
		})
	}
	encoded, err := json.MarshalIndent(states, "", "  ")
	if err != nil {
		return "", "", err
	}
	return string(encoded), statesToCSV(states), nil
}

func statesToCSV(states []exportedState) string {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	writer.UseCRLF = true
	writer.Write([]string{"key", "date", "title", "amount", "category_20", "category_20_confidence", "essential_probability"})
	for _, s := range states {
		writer.Write([]string{s.Key, s.Date, s.Title, string(s.Amount), s.Category20.Choice,
			strconv.FormatFloat(s.Category20.Confidence, 'f', 2, 64), strconv.FormatFloat(s.Essential.Noul, 'f', 2, 64)})
	}
	writer.Flush()
	return buffer.String()
}

func sortedByDate(classified []ClassifiedTransaction) ([]ClassifiedTransaction, error) {
	type dated struct {
		date        time.Time
		transaction ClassifiedTransaction
	}
	items := make([]dated, len(classified))
	for i, c := range classified {
		date, err := parseDate(c.Transaction.Date)
		if err != nil {
			return nil, err
		}
		items[i] = dated{date, c}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].date.Before(items[j].date) })
	ordered := make([]ClassifiedTransaction, len(items))
	for i, item := range items {
		ordered[i] = item.transaction
	}
	return ordered, nil
}

// Function Contract:
// Input: expenditure totals
// Output: JSON with expenditure per month, per category and overall, in dollars
func ExportTotals(totals Totals) (string, error) {
	export := totalsExport{
		ExpenditureByMonth:    map[string]json.Number{},
		ExpenditureByCategory: map[string]json.Number{},
		TotalExpenditure:      formatDollars(totals.Total),
	}
	for month, cents := range totals.ByMonth {
		export.ExpenditureByMonth[month] = formatDollars(cents)
	}
	for category, cents := range totals.ByCategory {
		export.ExpenditureByCategory[category] = formatDollars(cents)
	}
	encoded, err := json.MarshalIndent(export, "", "  ")
	return string(encoded), err
}

const keyFieldSeparator = "\x1f" // unit separator: cannot appear in a title, so fields never run together

// Function Contract:
// Input: a transaction
// Output: the hex SHA-256 of its (date, title, amount) triple
// Logic: Computer normalizes the date to MM/DD/YY, joins date, title and amount in cents, and hashes them.
// Two different transactions with the same date, title and amount (two identical purchases on one day)
// share a key by design of the triple.
func TransactionKey(transaction Transaction) (string, error) {
	parsed, err := parseDate(transaction.Date)
	if err != nil {
		return "", err
	}
	triple := parsed.Format(statementDateLayout) + keyFieldSeparator + transaction.Description + keyFieldSeparator + strconv.FormatInt(transaction.AmountCents, 10)
	sum := sha256.Sum256([]byte(triple))
	return hex.EncodeToString(sum[:]), nil
}

// Totals are expenditure in cents, positive for money that left the account. Months are keyed "YYYY-MM".
type Totals struct {
	ByMonth    map[string]int64
	ByCategory map[string]int64
	Total      int64
}

// Function Contract:
// Input: final decisions
// Output: expenditure per (year, month), per category, and overall
// Logic: Computer sums the negated amount of every outflow (a negative amount), grouped by the month
// of its date and by its category. Positive amounts are not expenditure and are left out.
func ComputeTotals(decisions []Decision) (Totals, error) {
	totals := Totals{ByMonth: map[string]int64{}, ByCategory: map[string]int64{}}
	for _, decision := range decisions {
		if decision.AmountCents >= 0 {
			continue
		}
		date, err := parseDate(decision.Date)
		if err != nil {
			return Totals{}, err
		}
		spend := -decision.AmountCents
		totals.ByMonth[date.Format("2006-01")] += spend
		totals.ByCategory[decision.Category] += spend
		totals.Total += spend
	}
	return totals, nil
}
