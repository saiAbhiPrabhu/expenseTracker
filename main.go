// main.go

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const categoriesFile = "transaction_categories.json"

func main() {
	abTestDir := flag.String("abtest-dir", "", "run the JEV A/B test and write results.json and report.md here")
	artifactsDir := flag.String("build-demo-artifacts", "", "read the statement PDFs given as arguments and write the demo's saved Process artifacts to this directory")
	var ocrFixes []OCRFix
	serveAddr := flag.String("serve", "", "start the web app on this address (for example :8080)")
	webDir := flag.String("web-dir", "web/dist", "the built front end that -serve serves")
	exportDir := flag.String("export", "", "read the A/B results.json files given as arguments and write the combined states and expenditure totals to this directory")
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
	if *artifactsDir != "" {
		if err := BuildDemoArtifacts(*artifactsDir, flag.Args()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *serveAddr != "" {
		demo, err := loadDemoFiles()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "serving %s and the demo API on %s\n", *webDir, *serveAddr)
		if err := http.ListenAndServe(*serveAddr, NewWebHandler(os.DirFS(*webDir), demo)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *exportDir != "" {
		if err := runExport(flag.Args(), *exportDir); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
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

// loadSpending returns the scrubbed transactions to classify: the titles from titlesCSV when given
// (already extracted), otherwise the spending rows of the statement PDF.
func loadSpending(pdfFile string, snapshotDir string, ocrFixes []OCRFix, titlesCSV string) ([]Transaction, error) {
	if titlesCSV != "" {
		data, err := os.ReadFile(titlesCSV)
		if err != nil {
			return nil, err
		}
		transactions, err := ParseTitlesCSV(string(data))
		if err != nil {
			return nil, err
		}
		return ScrubTransactions(transactions), nil // scrubbing is idempotent, so scrubbed titles are unchanged
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

// runExport reads the results.json files of several A/B runs and writes the export files to dir.
// It makes no API calls; the content comes from Export.
func runExport(resultFiles []string, dir string) error {
	var classified []ClassifiedTransaction
	for _, file := range resultFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		var results []ClassifiedTransaction
		if err := json.Unmarshal(data, &results); err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
		classified = append(classified, results...)
	}
	files, err := Export(classified)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for name, content := range map[string]string{"jev_states.json": files.StatesJSON, "jev_states.csv": files.StatesCSV, "expenditure_totals.json": files.TotalsJSON} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// loadDemoFiles reads the saved demo run (see demo/cache) and the category list.
func loadDemoFiles() (*DemoData, error) {
	results, err := os.ReadFile(filepath.Join("demo", "cache", "results.json"))
	if err != nil {
		return nil, fmt.Errorf("demo data missing (run the demo cache first, see README): %w", err)
	}
	categories, err := os.ReadFile(categoriesFile)
	if err != nil {
		return nil, err
	}
	return LoadDemoData(results, categories)
}
