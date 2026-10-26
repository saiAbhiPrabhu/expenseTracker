// demo_artifacts.go
//
// Responsibility: the builder that reads the sample statements through the real pipeline and saves what each
// stage produced (demo/artifacts/<id>/process.json and the page images), so serving the demo needs no OCR.
// Run it with: go run . -build-demo-artifacts demo/artifacts demo/statements/*.pdf

package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	displayDPI        = 90 // the page images shown on screen; the pipeline itself reads at renderDPI
	summaryPageReason = "The first page is the account summary. It is never read."
	blankPageReason   = "No transactions on this page."
)

type pageReading struct {
	Page              int     `json:"page"`
	Status            string  `json:"status"`
	Reason            string  `json:"reason,omitempty"`
	BeforeImage       string  `json:"beforeImage,omitempty"`
	AfterImage        string  `json:"afterImage,omitempty"`
	RemovedTopPercent float64 `json:"removedTopPercent,omitempty"`
	OCRText           string  `json:"ocrText,omitempty"`
}

type sectionCheck struct {
	Name         string      `json:"name"`
	Rows         int         `json:"rows"`
	PrintedTotal json.Number `json:"printedTotal"`
	RowsTotal    json.Number `json:"rowsTotal"`
	Matches      bool        `json:"matches"`
}

type processRow struct {
	Date          string      `json:"date"`
	Amount        json.Number `json:"amount"`
	Title         string      `json:"title"`
	ScrubbedTitle string      `json:"scrubbedTitle"`
	Section       string      `json:"section"`
	Kept          bool        `json:"kept"`
}

// processSummary: FlaggedForReview is left out of the saved file; the server fills it in per request.
type processSummary struct {
	RowsFound        int  `json:"rowsFound"`
	DepositsRemoved  int  `json:"depositsRemoved"`
	SpendingRows     int  `json:"spendingRows"`
	FlaggedForReview *int `json:"flaggedForReview,omitempty"`
}

type statementProcess struct {
	ID           string         `json:"id"`
	Label        string         `json:"label"`
	FileName     string         `json:"fileName"`
	Pages        int            `json:"pages"`
	PageReadings []pageReading  `json:"pageReadings"`
	Sections     []sectionCheck `json:"sections"`
	Rows         []processRow   `json:"rows"`
	Summary      processSummary `json:"summary"`
}

// pageScan is what the pipeline found on one page. TopPixels and HeightPixels are at renderDPI.
type pageScan struct {
	Page         int
	Found        bool // a red heading was found, so the page was cropped and read
	TopPixels    int
	HeightPixels int
	OCRText      string
}

// scannedPage adds the display images, which are only written to disk, never put in the JSON.
type scannedPage struct {
	pageScan
	before, after image.Image
}

var statementFileName = regexp.MustCompile(`^statement-(\d{4}-\d{2})\.pdf$`)

// Function Contract:
// Input: the id (YYYY-MM), the pdf's file name, and what the pipeline found on each page (page 1 first)
// Output: the saved process record
// Logic: Computer joins the text of the read pages, extracts and scrubs the rows, and tags deposits as removed.
func buildProcess(id string, fileName string, scans []pageScan) (statementProcess, error) {
	month, err := time.Parse("2006-01", id)
	if err != nil {
		return statementProcess{}, err
	}
	var text strings.Builder
	for _, scan := range scans {
		text.WriteString(scan.OCRText)
	}
	transactions, sections, err := ExtractWithSections(text.String())
	if err != nil {
		return statementProcess{}, fmt.Errorf("%s: %w", id, err)
	}
	rows := buildRows(transactions, ScrubTransactions(transactions))
	spending := len(RemoveDeposits(transactions))
	return statementProcess{
		ID: id, Label: month.Format("January 2006"), FileName: fileName, Pages: len(scans),
		PageReadings: buildPageReadings(id, scans),
		Sections:     buildSectionChecks(sections),
		Rows:         rows,
		Summary:      processSummary{RowsFound: len(rows), DepositsRemoved: len(rows) - spending, SpendingRows: spending},
	}, nil
}

func buildPageReadings(id string, scans []pageScan) []pageReading {
	readings := make([]pageReading, len(scans))
	for i, scan := range scans {
		switch {
		case scan.Found:
			prefix := fmt.Sprintf("/api/demo/statements/%s/pages/%d/", id, scan.Page)
			readings[i] = pageReading{
				Page: scan.Page, Status: "read", BeforeImage: prefix + "before.png", AfterImage: prefix + "after.png",
				RemovedTopPercent: math.Round(1000*float64(scan.TopPixels)/float64(scan.HeightPixels)) / 10,
				OCRText:           scan.OCRText,
			}
		case scan.Page == 1:
			readings[i] = pageReading{Page: 1, Status: "skipped", Reason: summaryPageReason}
		default:
			readings[i] = pageReading{Page: scan.Page, Status: "skipped", Reason: blankPageReason}
		}
	}
	return readings
}

func buildSectionChecks(sections []SectionSummary) []sectionCheck {
	checks := make([]sectionCheck, len(sections))
	for i, s := range sections {
		checks[i] = sectionCheck{s.Name, s.Rows, formatDollars(s.PrintedCents), formatDollars(s.RowsCents), s.PrintedCents == s.RowsCents}
	}
	return checks
}

func buildRows(transactions, scrubbed []Transaction) []processRow {
	rows := make([]processRow, len(transactions))
	for i, t := range transactions {
		rows[i] = processRow{t.Date, formatDollars(t.AmountCents), t.Description, scrubbed[i].Description, t.Section, t.Section != depositsSection}
	}
	return rows
}

// BuildDemoArtifacts reads every statement pdf and writes outDir/<id>/process.json plus the before and
// after image of each read page. Needs pdftoppm, pdfinfo and tesseract.
func BuildDemoArtifacts(outDir string, pdfFiles []string) error {
	for _, pdfFile := range pdfFiles {
		match := statementFileName.FindStringSubmatch(filepath.Base(pdfFile))
		if match == nil {
			return fmt.Errorf("%s: want a file named statement-YYYY-MM.pdf", pdfFile)
		}
		if err := buildStatementArtifacts(filepath.Join(outDir, match[1]), match[1], pdfFile); err != nil {
			return err
		}
	}
	return nil
}

func buildStatementArtifacts(dir, id, pdfFile string) error {
	pages, err := scanStatement(pdfFile)
	if err != nil {
		return err
	}
	scans := make([]pageScan, len(pages))
	for i, page := range pages {
		scans[i] = page.pageScan
	}
	process, err := buildProcess(id, filepath.Base(pdfFile), scans)
	if err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(process, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, page := range pages {
		if !page.Found {
			continue
		}
		if err := writePNG(filepath.Join(dir, fmt.Sprintf("page-%d-before.png", page.Page)), page.before); err != nil {
			return err
		}
		if err := writePNG(filepath.Join(dir, fmt.Sprintf("page-%d-after.png", page.Page)), page.after); err != nil {
			return err
		}
	}
	return os.WriteFile(filepath.Join(dir, "process.json"), encoded, 0o644)
}

// scanStatement runs the pipeline's page steps (render, crop above the first red heading, OCR) and keeps
// the numbers and images of each. Page 1 is never rendered.
func scanStatement(pdfFile string) ([]scannedPage, error) {
	workDir, err := os.MkdirTemp("", "statement-artifacts-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(workDir)
	pageCount, err := countPages(pdfFile)
	if err != nil {
		return nil, err
	}
	pages := make([]scannedPage, pageCount)
	for n := 1; n <= pageCount; n++ {
		pages[n-1].Page = n
		if n < firstTransactionPage {
			continue
		}
		pages[n-1], err = scanPage(pdfFile, n, workDir)
		if err != nil {
			return nil, err
		}
	}
	return pages, nil
}

func scanPage(pdfFile string, n int, workDir string) (scannedPage, error) {
	full, err := renderPage(pdfFile, n, workDir)
	if err != nil {
		return scannedPage{}, err
	}
	cropped, found := cropAboveFirstRedText(full)
	if !found {
		return scannedPage{pageScan: pageScan{Page: n}}, nil
	}
	text, err := ocrImage(cropped, workDir)
	if err != nil {
		return scannedPage{}, err
	}
	top := cropped.Bounds().Min.Y - full.Bounds().Min.Y
	before, err := renderPageAt(pdfFile, n, displayDPI, workDir)
	if err != nil {
		return scannedPage{}, err
	}
	return scannedPage{
		pageScan: pageScan{Page: n, Found: true, TopPixels: top, HeightPixels: full.Bounds().Dy(), OCRText: text},
		before:   before,
		after:    cropTop(before, top*displayDPI/renderDPI),
	}, nil
}

// cropTop removes the top rows of an image.
func cropTop(img image.Image, rows int) image.Image {
	b := img.Bounds()
	return img.(interface {
		SubImage(image.Rectangle) image.Image
	}).SubImage(image.Rect(b.Min.X, b.Min.Y+rows, b.Max.X, b.Max.Y))
}

func renderPageAt(pdfFile string, page, dpi int, workDir string) (image.Image, error) {
	prefix := filepath.Join(workDir, fmt.Sprintf("display-%d", page))
	pageArg := strconv.Itoa(page)
	err := exec.Command("pdftoppm", "-r", strconv.Itoa(dpi), "-f", pageArg, "-l", pageArg,
		"-png", "-singlefile", pdfFile, prefix).Run()
	if err != nil {
		return nil, fmt.Errorf("pdftoppm page %d: %w", page, err)
	}
	file, err := os.Open(prefix + ".png")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return png.Decode(file)
}
