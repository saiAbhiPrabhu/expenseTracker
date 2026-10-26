// demo_artifacts_test.go

package main

import (
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func sampleScans() []pageScan {
	return []pageScan{
		{Page: 1},
		{Page: 2},
		{Page: 3, Found: true, TopPixels: 100, HeightPixels: 800, OCRText: sampleText},
	}
}

func TestBuildPageReadings(t *testing.T) {
	got := buildPageReadings("2026-06", sampleScans())
	if len(got) != 3 {
		t.Fatalf("got %d readings, want 3", len(got))
	}
	if got[0].Status != "skipped" || got[0].Reason != summaryPageReason {
		t.Errorf("page 1: %+v", got[0])
	}
	if got[1].Status != "skipped" || got[1].Reason != blankPageReason {
		t.Errorf("page 2: %+v", got[1])
	}
	want := pageReading{
		Page: 3, Status: "read", BeforeImage: "/api/demo/statements/2026-06/pages/3/before.png",
		AfterImage: "/api/demo/statements/2026-06/pages/3/after.png", RemovedTopPercent: 12.5, OCRText: sampleText,
	}
	if got[2] != want {
		t.Errorf("page 3: got %+v, want %+v", got[2], want)
	}
}

func TestBuildProcess(t *testing.T) {
	got, err := buildProcess("2026-06", "statement-2026-06.pdf", sampleScans())
	if err != nil {
		t.Fatal(err)
	}
	if got.Label != "June 2026" || got.Pages != 3 {
		t.Errorf("label/pages: %q %d", got.Label, got.Pages)
	}
	if want := (processSummary{RowsFound: 4, DepositsRemoved: 2, SpendingRows: 2}); got.Summary != want {
		t.Errorf("summary: got %+v, want %+v", got.Summary, want)
	}
	if len(got.Sections) != 2 || !got.Sections[1].Matches || got.Sections[1].RowsTotal != "-24.25" {
		t.Errorf("sections: %+v", got.Sections)
	}
	if row := got.Rows[2]; row.Kept != true || row.Title != "COFFEE SHOP ID:123 PPD" || row.ScrubbedTitle != "COFFEE SHOP ID:[REDACTED]" {
		t.Errorf("row 3: %+v", row)
	}
	if got.Rows[0].Kept {
		t.Errorf("a deposit must not be kept: %+v", got.Rows[0])
	}
}

func TestBuildProcessChecksumMismatch(t *testing.T) {
	scans := []pageScan{{Page: 1}, {Page: 2, Found: true, HeightPixels: 1, OCRText: "Deposits\nDate Description Amount\n01/02/25 PAY 10.00\nTotal deposits $11.00\n"}}
	if _, err := buildProcess("2026-06", "x.pdf", scans); err == nil {
		t.Fatal("expected checksum error")
	}
}

func TestCropTop(t *testing.T) {
	got := cropTop(image.NewRGBA(image.Rect(0, 0, 10, 20)), 5)
	if got.Bounds().Dy() != 15 || got.Bounds().Dx() != 10 {
		t.Errorf("got %v", got.Bounds())
	}
}

// TestBuildDemoArtifactsRealPDF builds one statement's artifacts and checks the saved JSON matches
// the committed artifacts.
func TestBuildDemoArtifactsRealPDF(t *testing.T) {
	if testing.Short() {
		t.Skip("OCR is slow")
	}
	for _, tool := range []string{"tesseract", "pdftoppm", "pdfinfo"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not on PATH", tool)
		}
	}
	dir := t.TempDir()
	if err := BuildDemoArtifacts(dir, []string{"demo/statements/statement-2026-06.pdf"}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "2026-06", "process.json"))
	if err != nil {
		t.Fatal(err)
	}
	committed, err := os.ReadFile("demo/artifacts/2026-06/process.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(committed) {
		t.Error("demo/artifacts/2026-06/process.json is stale; rebuild with -build-demo-artifacts")
	}
	for _, secret := range []string{"JAMIE", "EXAMPLE", "1234 5678", "Account #"} {
		if strings.Contains(string(got), secret) {
			t.Errorf("process.json contains %q", secret)
		}
	}
	for _, name := range []string{"page-3-before.png", "page-3-after.png", "page-5-after.png"} {
		if _, err := os.Stat(filepath.Join(dir, "2026-06", name)); err != nil {
			t.Error(err)
		}
	}
}

func TestBuildDemoArtifactsRejectsOddFileName(t *testing.T) {
	if err := BuildDemoArtifacts(t.TempDir(), []string{"whatever.pdf"}); err == nil {
		t.Fatal("expected an error")
	}
}
