package main

import "testing"

const sampleText = `Deposits and other additions
Date Description Amount
01/02/25 PAYROLL DEPOSIT 1,000.00
01/03/25 REFUND 5.50
Total deposits and other additions $1,005.50
Withdrawals and other subtractions
ATM and debit card subtractions
Date Description Amount
01/04/25 COFFEE SHOP -4.25
ID:123 PPD
continued on the next page
ATM and debit card subtractions - continued
Date Description Amount
01/05/25 GROCERY -20.00
Total ATM and debit card subtractions -$24.25
Page 3 of 10
`

func TestExtractTransactionsFromText(t *testing.T) {
	got, err := ExtractTransactionsFromText(sampleText)
	if err != nil {
		t.Fatal(err)
	}
	want := []Transaction{
		{"01/02/25", "PAYROLL DEPOSIT", 100000, "Deposits and other additions"},
		{"01/03/25", "REFUND", 550, "Deposits and other additions"},
		{"01/04/25", "COFFEE SHOP ID:123 PPD", -425, "ATM and debit card subtractions"},
		{"01/05/25", "GROCERY", -2000, "ATM and debit card subtractions"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d rows, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestExtractTransactionsFromTextChecksumMismatch(t *testing.T) {
	_, err := ExtractTransactionsFromText("Deposits\nDate Description Amount\n01/02/25 PAY 10.00\nTotal deposits $11.00\n")
	if err == nil {
		t.Fatal("expected checksum error")
	}
}

func TestConvertTransactionsToJSON(t *testing.T) {
	got, err := ConvertTransactionsToJSON([]Transaction{
		{"01/04/25", "COFFEE SHOP", -425, "x"},
		{"2025-01-05", "PAYROLL", 100000, "y"},
		{"1/6/2025", "TOTAL", -164160, "z"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := `[{"Date":"01/04/25","Description":"COFFEE SHOP","Amount":-4.25},` +
		`{"Date":"01/05/25","Description":"PAYROLL","Amount":1000.00},` +
		`{"Date":"01/06/25","Description":"TOTAL","Amount":-1641.60}]`
	if got != want {
		t.Errorf("got %s\nwant %s", got, want)
	}
}

func TestConvertTransactionsToJSONBadDate(t *testing.T) {
	if _, err := ConvertTransactionsToJSON([]Transaction{{"yesterday", "X", 1, ""}}); err == nil {
		t.Fatal("expected date error")
	}
}
