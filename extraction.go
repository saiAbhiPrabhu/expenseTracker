// extraction.go

package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Extraction: turns statement text into transactions (grab the features, standardize them), anonymizes
// them, and constructs the JSON states for JEV.

// Transaction is one statement row. Amount is in cents, negative for withdrawals.
type Transaction struct {
	Date        string
	Description string
	AmountCents int64
	Section     string
}

var (
	tableHeaderLine  = "Date Description Amount"
	pageEndLine      = "continued on the next page"
	continuedSuffix  = " - continued"
	datePrefix       = regexp.MustCompile(`^\d{2}/\d{2}/\d{2}\b`)
	transactionRow   = regexp.MustCompile(`^(\d{2}/\d{2}/\d{2})\s+(.*\S)\s+(-?[\d,]+\.\d{2})$`)
	sectionTotalLine = regexp.MustCompile(`^Total\b.*\s(-?)\$([\d,]+\.\d{2})$`)
)

// Function Contract:
// Input: textFile as a string
// Output: list of transactions
// Logic: Computer reads the OCR text and returns every transaction row, tagged with its section.
// Each section's printed Total is checked against the sum of its rows; a mismatch means OCR misread a row.
func ExtractTransactionsFromText(textFile string) ([]Transaction, error) {
	transactions, _, err := ExtractWithSections(textFile)
	return transactions, err
}

// SectionSummary is the check of one section: its row count, the Total the statement prints and the
// sum of the rows. They always match when extraction succeeds.
type SectionSummary struct {
	Name         string
	Rows         int
	PrintedCents int64
	RowsCents    int64
}

// ExtractWithSections is ExtractTransactionsFromText that also reports each section's check, in the
// order the statement prints the section Totals.
func ExtractWithSections(textFile string) ([]Transaction, []SectionSummary, error) {
	var (
		transactions []Transaction
		sections     []SectionSummary
		sectionSums  = map[string]int64{}
		sectionRows  = map[string]int{}
		section      string
		capturing    bool
		previousLine string
	)
	for _, rawLine := range strings.Split(textFile, "\n") {
		line := strings.TrimSpace(rawLine)
		switch {
		case line == tableHeaderLine:
			section = strings.TrimSuffix(previousLine, continuedSuffix)
			capturing = true
		case !capturing || line == "":
		case line == pageEndLine:
			capturing = false
		case sectionTotalLine.MatchString(line):
			capturing = false
			if err := checkSectionTotal(section, line, sectionSums[section]); err != nil {
				return nil, nil, err
			}
			printed, _ := printedTotalCents(line) // already parsed by checkSectionTotal
			sections = append(sections, SectionSummary{section, sectionRows[section], printed, sectionSums[section]})
		case datePrefix.MatchString(line):
			transaction, err := parseTransactionRow(line, section)
			if err != nil {
				return nil, nil, err
			}
			transactions = append(transactions, transaction)
			sectionSums[section] += transaction.AmountCents
			sectionRows[section]++
		case len(transactions) > 0:
			last := &transactions[len(transactions)-1]
			last.Description += " " + line
		}
		if line != "" {
			previousLine = line
		}
	}
	return transactions, sections, nil
}

const depositsSection = "Deposits and other additions"

// RemoveDeposits drops the money-in rows. It runs after ExtractTransactionsFromText because the
// deposits rows are still needed there to check the printed Total.
func RemoveDeposits(transactions []Transaction) []Transaction {
	var spending []Transaction
	for _, transaction := range transactions {
		if transaction.Section != depositsSection {
			spending = append(spending, transaction)
		}
	}
	return spending
}

func parseTransactionRow(line string, section string) (Transaction, error) {
	match := transactionRow.FindStringSubmatch(line)
	if match == nil {
		return Transaction{}, fmt.Errorf("row starts with a date but has no amount: %q", line)
	}
	cents, err := parseCents(match[3])
	if err != nil {
		return Transaction{}, err
	}
	return Transaction{Date: match[1], Description: match[2], AmountCents: cents, Section: section}, nil
}

func checkSectionTotal(section string, totalLine string, rowSumCents int64) error {
	printed, err := printedTotalCents(totalLine)
	if err != nil {
		return err
	}
	if printed != rowSumCents {
		return fmt.Errorf("section %q: printed total %d cents, rows sum to %d cents", section, printed, rowSumCents)
	}
	return nil
}

func printedTotalCents(totalLine string) (int64, error) {
	match := sectionTotalLine.FindStringSubmatch(totalLine)
	return parseCents(match[1] + match[2])
}

// parseCents converts "-1,641.60" to -164160.
func parseCents(amount string) (int64, error) {
	digits := strings.NewReplacer(",", "", ".", "").Replace(amount)
	return strconv.ParseInt(digits, 10, 64)
}

// jevState is the JSON object JEV receives for one transaction.
type jevState struct {
	Date        string      `json:"Date"`
	Description string      `json:"Description"`
	Amount      json.Number `json:"Amount"`
}

const statementDateLayout = "01/02/06"

// Other date spellings that are rewritten to statementDateLayout.
var alternateDateLayouts = []string{"2006-01-02", "1/2/06", "1/2/2006", "01/02/2006"}

// Function Contract:
// Input: list of transactions
// Output: JSON array of JEV states as a string
// Logic: Computer converts each transaction to {"Date", "Description", "Amount"}, with the date
// enforced as MM/DD/YY and the amount as a number in dollars.
func ConvertTransactionsToJSON(transactions []Transaction) (string, error) {
	states := make([]jevState, 0, len(transactions))
	for _, transaction := range transactions {
		state, err := toJEVState(transaction)
		if err != nil {
			return "", err
		}
		states = append(states, state)
	}
	encoded, err := json.Marshal(states)
	return string(encoded), err
}

func toJEVState(transaction Transaction) (jevState, error) {
	date, err := parseDate(transaction.Date)
	if err != nil {
		return jevState{}, err
	}
	return jevState{
		Date:        date.Format(statementDateLayout),
		Description: transaction.Description,
		Amount:      formatDollars(transaction.AmountCents),
	}, nil
}

// parseDate reads a date written as MM/DD/YY or one of the alternate spellings. Every date in the
// code goes through here, so the accepted formats are defined in one place.
func parseDate(date string) (time.Time, error) {
	for _, layout := range append([]string{statementDateLayout}, alternateDateLayouts...) {
		if parsed, err := time.Parse(layout, date); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized date %q", date)
}

// formatDollars converts -164160 to the JSON number -1641.60.
func formatDollars(cents int64) json.Number {
	sign := ""
	if cents < 0 {
		sign, cents = "-", -cents
	}
	return json.Number(fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100))
}

const redacted = "[REDACTED]"

const (
	peerSeparator = `[^A-Za-z0-9\[]+` // anything but a letter, digit or the "[" of "[REDACTED]"
	peerCue       = peerSeparator + `(?:payment|transfer|to|from|sent|received)`
)

var (
	// A peer-payment keyword and optional cue words, split by any punctuation OCR produces (VENMO?*);
	// everything after them is the counterparty and memo. PayPal is mostly a processor in front of a
	// merchant ("PAYPAL *WOK EXPRESS"), so it only counts when a cue word shows a person is paid
	// ("PAYPAL PAYMENT TO ...").
	peerParty = regexp.MustCompile(`(?i)(\b(?:(?:zelle|venmo|cash ?app)\b(?:` + peerCue + `)*|paypal\b(?:` + peerCue + `)+)` + peerSeparator + `).+$`)
	// Any "KEY:" label; the text after it, up to the next label, is that field's value.
	fieldLabel = regexp.MustCompile(`[A-Za-z_]+:`)
	// Account, card and reference numbers. Shorter runs are kept: statement dates are 4 digits.
	longNumber = regexp.MustCompile(`\d{6,}`)
)

// Function Contract:
// Input: transactions
// Output: copies with anonymized descriptions
// Logic: Computer masks everything after Zelle/Venmo-style keywords and every KEY:value field value,
// then masks long digit runs.
func ScrubTransactions(transactions []Transaction) []Transaction {
	scrubbed := make([]Transaction, len(transactions))
	for i, transaction := range transactions {
		transaction.Description = scrubDescription(transaction.Description)
		scrubbed[i] = transaction
	}
	return scrubbed
}

func scrubDescription(description string) string {
	description = peerParty.ReplaceAllString(description, "${1}"+redacted)
	description = maskFieldValues(description)
	return longNumber.ReplaceAllString(description, redacted)
}

// maskFieldValues keeps each label and replaces its value.
func maskFieldValues(description string) string {
	labels := fieldLabel.FindAllStringIndex(description, -1)
	if labels == nil {
		return description
	}
	var out strings.Builder
	out.WriteString(description[:labels[0][1]])
	for i, label := range labels {
		valueEnd := len(description)
		if i+1 < len(labels) {
			valueEnd = labels[i+1][0]
		}
		out.WriteString(redactKeepingSpaces(description[label[1]:valueEnd]))
		if i+1 < len(labels) {
			out.WriteString(description[labels[i+1][0]:labels[i+1][1]])
		}
	}
	return out.String()
}

func redactKeepingSpaces(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return value
	}
	start := strings.Index(value, trimmed)
	return value[:start] + redacted + value[start+len(trimmed):]
}
