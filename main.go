// main.go

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const categoriesFile = "transaction_categories.json"

func main() {
	abTestDir := flag.String("abtest-dir", "", "run the JEV A/B test and write results.json and report.md here")
	flag.Parse()
	if err := run(flag.Arg(0), os.Getenv("SNAPSHOT_DIR"), *abTestDir, os.Getenv("TYPESAFE_API_KEY")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(pdfFile string, snapshotDir string, abTestDir string, apiKey string) error {
	text, err := ConvertPDFToText(pdfFile, snapshotDir)
	if err != nil {
		return err
	}
	transactions, err := ExtractTransactionsFromText(text)
	if err != nil {
		return err
	}
	spending := ScrubTransactions(RemoveDeposits(transactions))
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
