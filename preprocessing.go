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
	"regexp"
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
	var (
		transactions []Transaction
		sectionSums  = map[string]int64{}
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
				return nil, err
			}
		case datePrefix.MatchString(line):
			transaction, err := parseTransactionRow(line, section)
			if err != nil {
				return nil, err
			}
			transactions = append(transactions, transaction)
			sectionSums[section] += transaction.AmountCents
		case len(transactions) > 0:
			last := &transactions[len(transactions)-1]
			last.Description += " " + line
		}
		if line != "" {
			previousLine = line
		}
	}
	return transactions, nil
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
	match := sectionTotalLine.FindStringSubmatch(totalLine)
	printed, err := parseCents(match[1] + match[2])
	if err != nil {
		return err
	}
	if printed != rowSumCents {
		return fmt.Errorf("section %q: printed total %d cents, rows sum to %d cents", section, printed, rowSumCents)
	}
	return nil
}

// parseCents converts "-1,641.60" to -164160.
func parseCents(amount string) (int64, error) {
	digits := strings.NewReplacer(",", "", ".", "").Replace(amount)
	return strconv.ParseInt(digits, 10, 64)
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
	transactions, err := ExtractTransactionsFromText(text)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, t := range transactions {
		fmt.Printf("%s | %s | %d | %s\n", t.Date, t.Section, t.AmountCents, t.Description)
	}
}
