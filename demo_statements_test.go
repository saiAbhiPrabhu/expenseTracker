// demo_statements_test.go

package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestDemoStatementsRoundTrip runs the real OCR pipeline over the synthetic demo statements and
// checks that every row of demo/transactions.csv comes back exactly.
func TestDemoStatementsRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("OCR is slow")
	}
	for _, tool := range []string{"tesseract", "pdftoppm"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not on PATH", tool)
		}
	}
	file, err := os.Open("demo/transactions.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	for _, month := range []string{"06", "07", "08", "09"} {
		t.Run(month, func(t *testing.T) {
			t.Parallel()
			var want []Transaction
			for _, record := range records[1:] {
				if record[0][:2] != month {
					continue
				}
				cents, err := parseCents(record[1])
				if err != nil {
					t.Fatal(err)
				}
				want = append(want, Transaction{Date: record[0], Description: record[2], AmountCents: cents})
			}
			text, err := ConvertPDFToText(fmt.Sprintf("demo/statements/statement-2026-%s.pdf", month), "")
			if err != nil {
				t.Fatal(err)
			}
			for _, secret := range []string{"JAMIE", "EXAMPLE", "1234 5678", "9012", "Account #"} {
				if strings.Contains(text, secret) {
					t.Errorf("OCR text contains %q", secret)
				}
			}
			got, err := ExtractTransactionsFromText(text)
			if err != nil {
				t.Fatal(err)
			}
			spending := RemoveDeposits(got)
			if deposits := len(got) - len(spending); deposits < 2 || deposits > 3 {
				t.Errorf("got %d deposits, want 2-3", deposits)
			}
			if len(spending) != len(want) {
				t.Fatalf("got %d spending rows, want %d", len(spending), len(want))
			}
			// Statement sections reorder rows relative to the CSV, so match each wanted row exactly.
			used := make([]bool, len(spending))
			for _, w := range want {
				found := false
				for i, g := range spending {
					if !used[i] && g.Date == w.Date && g.AmountCents == w.AmountCents && g.Description == w.Description {
						used[i], found = true, true
						break
					}
				}
				if !found {
					t.Errorf("row not extracted exactly: want %q on %s (%d cents)", w.Description, w.Date, w.AmountCents)
					for i, g := range spending {
						if !used[i] && g.Date == w.Date && g.AmountCents == w.AmountCents {
							t.Errorf("  nearest extracted: %q", g.Description)
						}
					}
				}
			}
		})
	}
}
