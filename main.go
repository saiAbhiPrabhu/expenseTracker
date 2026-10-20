// main.go

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const categoriesFile = "transaction_categories.json"

func main() {
	abTestDir := flag.String("abtest-dir", "", "run the JEV A/B test and write results.json and report.md here")
	var ocrFixes []OCRFix
	titlesCSV := flag.String("titles-csv", "", "TEMPORARY: use already-extracted titles from this CSV instead of a PDF")
	flag.Func("fix-ocr", "manual OCR correction as OLD=NEW (repeatable)", func(value string) error {
		old, replacement, found := strings.Cut(value, "=")
		if !found || old == "" {
			return fmt.Errorf("want OLD=NEW, got %q", value)
		}
		ocrFixes = append(ocrFixes, OCRFix{Old: old, New: replacement})
		return nil
	})
	flag.Parse()
	if err := run(flag.Arg(0), os.Getenv("SNAPSHOT_DIR"), *abTestDir, os.Getenv("TYPESAFE_API_KEY"), ocrFixes, *titlesCSV); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(pdfFile string, snapshotDir string, abTestDir string, apiKey string, ocrFixes []OCRFix, titlesCSV string) error {
	spending, err := loadSpending(pdfFile, snapshotDir, ocrFixes, titlesCSV)
	if err != nil {
		return err
	}
	if abTestDir != "" {
		return runABTest(spending, abTestDir, apiKey)
	}
	statesJSON, err := ConvertTransactionsToJSON(spending)
	if err != nil {
		return err
	}
	fmt.Println(statesJSON)
	return nil
}

// runABTest classifies every transaction with both category sets and writes the raw results and
// a comparison report. Partial results are written even when a call fails midway.
func runABTest(transactions []Transaction, dir string, apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("TYPESAFE_API_KEY is not set")
	}
	categoriesJSON, err := os.ReadFile(categoriesFile)
	if err != nil {
		return err
	}
	questions, err := BuildQuestionSet(categoriesJSON)
	if err != nil {
		return err
	}
	classified, classifyErr := ClassifyTransactions(context.Background(), NewJEVClient(apiKey), transactions, questions)
	if err := writeABTestOutput(dir, classified); err != nil {
		return err
	}
	return classifyErr
}

func writeABTestOutput(dir string, classified []ClassifiedTransaction) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	results, err := json.MarshalIndent(classified, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "results.json"), results, 0o644); err != nil {
		return err
	}
	report := FormatABReport(CompareCategorySets(classified))
	return os.WriteFile(filepath.Join(dir, "report.md"), []byte(report), 0o644)
}

// loadSpending returns the transactions to classify: the titles from titlesCSV when given (already
// extracted and scrubbed), otherwise the scrubbed spending rows of the statement PDF.
func loadSpending(pdfFile string, snapshotDir string, ocrFixes []OCRFix, titlesCSV string) ([]Transaction, error) {
	if titlesCSV != "" {
		data, err := os.ReadFile(titlesCSV)
		if err != nil {
			return nil, err
		}
		return ParseTitlesCSV(string(data))
	}
	text, err := ConvertPDFToText(pdfFile, snapshotDir)
	if err != nil {
		return nil, err
	}
	text, err = ApplyOCRFixes(text, ocrFixes)
	if err != nil {
		return nil, err
	}
	transactions, err := ExtractTransactionsFromText(text)
	if err != nil {
		return nil, err
	}
	return ScrubTransactions(RemoveDeposits(transactions)), nil
}
