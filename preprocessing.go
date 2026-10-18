// preprocessing.go

package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// This file will contain the preprocessing steps to convert a bank statement pdf into an extracted list of transactions

// 1. Convert the pdf into a text file
// 2. Extract the transactions from the text file
// 3. Convert the transactions into a list of transactions
// 4. Save the list of transactions to a file

// Function Contract:
// Input: pdfFile on computer
// Output: textFile as a string
// Logic: Computer reads the pdf file and converts it to a text file and returns the text file as a string
func ConvertPDFToText(pdfFile string, snapshotDir string) (string, error) {
	// snapshotDir: when non-empty, every stage of every page is saved there for inspection
	// Step 1: Crop pdf file to only include transactions
	// Inter-Page crop: 2-end
	// Intra-Page crop: crop from the top of the page untill the first red text
	//
	// The top of each page holds the (sensitive) name and account number, always above
	// the first red heading. Pages are rendered to images, cropped, and only then OCR'd,
	// so the sensitive region is never turned into text.
	//
	// Step 2: Convert the cropped pdf file to a text file (OCR; the embedded text layer is corrupted)

	workDir, err := os.MkdirTemp("", "statement-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(workDir)

	pageCount, err := countPages(pdfFile)
	if err != nil {
		return "", err
	}

	var textFile strings.Builder
	for page := firstTransactionPage; page <= pageCount; page++ {
		pageImage, err := renderPage(pdfFile, page, workDir)
		if err != nil {
			return "", err
		}
		if err := saveSnapshotImage(snapshotDir, page, "1-render", pageImage); err != nil {
			return "", err
		}
		cropped, ok := cropAboveFirstRedText(pageImage)
		if !ok {
			continue // no red heading: nothing but the sensitive header
		}
		if err := saveSnapshotImage(snapshotDir, page, "2-cropped", cropped); err != nil {
			return "", err
		}
		pageText, err := ocrImage(cropped, workDir)
		if err != nil {
			return "", err
		}
		if err := saveSnapshotText(snapshotDir, page, "3-ocr", pageText); err != nil {
			return "", err
		}
		textFile.WriteString(pageText)
	}

	// return the text file as a string
	return textFile.String(), nil
}

const (
	firstTransactionPage = 2
	renderDPI            = 300
	// red headings start at the left margin; the red logo flag does not
	leftMarginPoints  = 60
	redRunPixels      = 5
	cropPaddingPoints = 4
)

func isRed(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	return r>>8 > 180 && g>>8 < 90 && b>>8 < 110
}

// cropAboveFirstRedText drops everything above the first red heading in the left margin.
// ok is false when the page has no such heading.
func cropAboveFirstRedText(page image.Image) (cropped image.Image, ok bool) {
	bounds := page.Bounds()
	marginPixels := leftMarginPoints * renderDPI / 72
	paddingPixels := cropPaddingPoints * renderDPI / 72
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		redPixels := 0
		for x := bounds.Min.X; x < bounds.Min.X+marginPixels && x < bounds.Max.X; x++ {
			if isRed(page.At(x, y)) {
				redPixels++
			}
		}
		if redPixels >= redRunPixels {
			top := max(bounds.Min.Y, y-paddingPixels)
			return page.(interface {
				SubImage(image.Rectangle) image.Image
			}).
				SubImage(image.Rect(bounds.Min.X, top, bounds.Max.X, bounds.Max.Y)), true
		}
	}
	return nil, false
}

func countPages(pdfFile string) (int, error) {
	out, err := exec.Command("pdfinfo", pdfFile).Output()
	if err != nil {
		return 0, fmt.Errorf("pdfinfo: %w", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if rest, found := strings.CutPrefix(line, "Pages:"); found {
			return strconv.Atoi(strings.TrimSpace(rest))
		}
	}
	return 0, fmt.Errorf("pdfinfo reported no page count for %s", pdfFile)
}

func renderPage(pdfFile string, page int, workDir string) (image.Image, error) {
	prefix := filepath.Join(workDir, fmt.Sprintf("page-%d", page))
	pageArg := strconv.Itoa(page)
	err := exec.Command("pdftoppm", "-r", strconv.Itoa(renderDPI), "-f", pageArg, "-l", pageArg,
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

func writePNG(path string, img image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}

func saveSnapshotImage(snapshotDir string, page int, stage string, img image.Image) error {
	if snapshotDir == "" {
		return nil
	}
	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		return err
	}
	return writePNG(filepath.Join(snapshotDir, fmt.Sprintf("page-%d-%s.png", page, stage)), img)
}

func saveSnapshotText(snapshotDir string, page int, stage string, text string) error {
	if snapshotDir == "" {
		return nil
	}
	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(snapshotDir, fmt.Sprintf("page-%d-%s.txt", page, stage)), []byte(text), 0o644)
}

func ocrImage(img image.Image, workDir string) (string, error) {
	imagePath := filepath.Join(workDir, "ocr-input.png")
	if err := writePNG(imagePath, img); err != nil {
		return "", err
	}
	// psm 6: one uniform block of text, which keeps each transaction row on one line
	out, err := exec.Command("tesseract", imagePath, "stdout", "--psm", "6").Output()
	if err != nil {
		return "", fmt.Errorf("tesseract: %w", err)
	}
	return string(out), nil
}

// Function Contract:
// Input: textFile as a string
// Output: list of transactions
// Logic: Computer reads the text file and extracts the transactions from the text file and returns the list of transactions
func ExtractTransactionsFromText() ([]string, error) {
	// Any row that starts with a date is a transaction
	// To extract each field in a transaction, we can delimit by tabs potentially
	// transaction fields are date, description, amount, balance
	// return the list of transactions
	return nil, errors.New("not implemented")
}

// Function Contract:
// Input: list of transactions
// Output: JSON file
// Logic: Computer reads the list of transactions and converts it to a JSON file
func ConvertTransactionsToJSON() (string, error) {
	// For each transaction, create a JSON object where each field is accompanied by a definition of the transaction (TBD)
	// Export the JSON file to the computer
	// return the JSON file as a string
	return "", errors.New("not implemented")
}

func main() {
	text, err := ConvertPDFToText(os.Args[1], os.Getenv("SNAPSHOT_DIR"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(text)
}
