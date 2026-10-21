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

func TestRemoveDeposits(t *testing.T) {
	got, err := ExtractTransactionsFromText(sampleText)
	if err != nil {
		t.Fatal(err)
	}
	spending := RemoveDeposits(got)
	if len(spending) != 2 {
		t.Fatalf("got %d rows, want the 2 card rows: %+v", len(spending), spending)
	}
	for _, transaction := range spending {
		if transaction.Section != "ATM and debit card subtractions" {
			t.Errorf("unexpected section kept: %+v", transaction)
		}
	}
}

func TestScrubDescription(t *testing.T) {
	cases := []struct{ input, want string }{
		{`Zelle payment to Jane Doe for "rent 9:30 pm"; Conf# abc123`, `Zelle payment to [REDACTED]`},
		{`Zelle payment from jane  DOE Conf# zzz`, `Zelle payment from [REDACTED]`},
		{`PURCHASE 0910 VENMO *Bob Smith 855-123 CA`, `PURCHASE 0910 VENMO *[REDACTED]`},
		{`PMNT SENT 0823 VENMO* Bob Smith Visa Direct NY`, `PMNT SENT 0823 VENMO* [REDACTED]`},
		{`PMNT SENT 0726 VENMO?* Jane Do Visa Direct NY`, `PMNT SENT 0726 VENMO?* [REDACTED]`},
		{`ACME LLC DES:PAYROLL ID:99 INDN:Jane Doe CO ID:12 WEB`, `ACME LLC DES:[REDACTED] ID:[REDACTED] INDN:[REDACTED] ID:[REDACTED]`},
		{`ACME LLC DES:X NAME:Jane Doe`, `ACME LLC DES:[REDACTED] NAME:[REDACTED]`},
		{`CHECKCARD 0910 STARBUCKS 12345678 CA`, `CHECKCARD 0910 STARBUCKS [REDACTED] CA`},
		{`Zelle 5551234`, `Zelle [REDACTED]`},
		{`CHECKCARD 0910 STARBUCKS CA`, `CHECKCARD 0910 STARBUCKS CA`},
	}
	for _, c := range cases {
		if got := scrubDescription(c.input); got != c.want {
			t.Errorf("%q\n got  %q\n want %q", c.input, got, c.want)
		}
	}
}

func TestScrubTransactionsKeepsOtherFields(t *testing.T) {
	in := []Transaction{{Date: "01/02/25", Description: "Zelle payment to Jane Doe", AmountCents: -500, Section: "S"}}
	out := ScrubTransactions(in)
	if out[0].Description != "Zelle payment to [REDACTED]" || out[0].AmountCents != -500 || out[0].Section != "S" {
		t.Errorf("got %+v", out[0])
	}
	if in[0].Description != "Zelle payment to Jane Doe" {
		t.Errorf("input mutated: %+v", in[0])
	}
}

func TestParseDate(t *testing.T) {
	for _, date := range []string{"01/06/25", "2025-01-06", "1/6/25", "1/6/2025", "01/06/2025"} {
		parsed, err := parseDate(date)
		if err != nil || parsed.Format(statementDateLayout) != "01/06/25" {
			t.Errorf("%q: got %v, %v", date, parsed, err)
		}
	}
	if _, err := parseDate("yesterday"); err == nil {
		t.Error("expected an error for an unrecognized date")
	}
}
