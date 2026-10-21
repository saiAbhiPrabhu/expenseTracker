package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestExportStates(t *testing.T) {
	classified := []ClassifiedTransaction{
		classifiedFor("05/02/26", "LATER", -500, "Groceries", 0.9, 0.8),
		classifiedFor("04/22/26", "EARLIER, WITH COMMA", -211, "Dining and Drinks", 0.5, 0.1),
	}
	classified[1].Category20.Probabilities = map[string]float64{"Dining and Drinks": 0.5, "Other": 0.5}
	statesJSON, statesCSV, err := ExportStates(classified)
	if err != nil {
		t.Fatal(err)
	}
	var states []exportedState
	if err := json.Unmarshal([]byte(statesJSON), &states); err != nil || len(states) != 2 {
		t.Fatalf("json: %v %s", err, statesJSON)
	}
	if states[0].Title != "EARLIER, WITH COMMA" || string(states[0].Amount) != "-2.11" || states[0].Category20.Probabilities["Other"] != 0.5 || len(states[0].Key) != 64 {
		t.Errorf("first state (should be the earlier date, with raw probabilities): %+v", states[0])
	}
	lines := strings.Split(strings.TrimSpace(statesCSV), "\r\n")
	if len(lines) != 3 || !strings.HasPrefix(lines[0], "key,date,title,amount,category_20,") {
		t.Fatalf("csv: %q", statesCSV)
	}
	if !strings.Contains(lines[1], `"EARLIER, WITH COMMA",-2.11,Dining and Drinks,0.50,0.10`) {
		t.Errorf("csv row: %q", lines[1])
	}
	if _, _, err := ExportStates([]ClassifiedTransaction{classifiedFor("someday", "X", -1, "Other", 1, 1)}); err == nil {
		t.Error("expected a date error")
	}
}

func TestExportTotals(t *testing.T) {
	got, err := ExportTotals(Totals{ByMonth: map[string]int64{"2026-04": 1250}, ByCategory: map[string]int64{"Groceries": 1250}, Total: 1250})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"2026-04": 12.50`, `"Groceries": 12.50`, `"total_expenditure": 12.50`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %s", want, got)
		}
	}
}

func TestTransactionKey(t *testing.T) {
	base := Transaction{Date: "04/22/26", Description: "DUNKIN", AmountCents: -211}
	key, err := TransactionKey(base)
	if err != nil || len(key) != 64 {
		t.Fatalf("got %q, %v", key, err)
	}
	if again, _ := TransactionKey(base); again != key {
		t.Error("key is not stable")
	}
	if iso, _ := TransactionKey(Transaction{Date: "2026-04-22", Description: "DUNKIN", AmountCents: -211}); iso != key {
		t.Error("the same date in another format must give the same key")
	}
	for name, other := range map[string]Transaction{
		"date":   {Date: "04/23/26", Description: "DUNKIN", AmountCents: -211},
		"title":  {Date: "04/22/26", Description: "DUNKIN 2", AmountCents: -211},
		"amount": {Date: "04/22/26", Description: "DUNKIN", AmountCents: -212},
		"joined": {Date: "04/22/26", Description: "DUNKIN\x1f-211", AmountCents: 0},
	} {
		if got, _ := TransactionKey(other); got == key {
			t.Errorf("changing the %s must change the key", name)
		}
	}
	if _, err := TransactionKey(Transaction{Date: "someday"}); err == nil {
		t.Error("expected an error for a bad date")
	}
}

func TestComputeTotals(t *testing.T) {
	decisions := []Decision{
		{Date: "04/22/26", Category: "Dining and Drinks", AmountCents: -1000},
		{Date: "04/30/26", Category: "Groceries", AmountCents: -250},
		{Date: "05/01/26", Category: "Dining and Drinks", AmountCents: -500},
		{Date: "05/02/26", Category: "Dining and Drinks", AmountCents: 9999},  // a credit is not expenditure
		{Date: "05/03/26", Category: "Income and Refunds", AmountCents: -300}, // an outflow counts, whatever its category
		{Date: "04/05/25", Category: "Groceries", AmountCents: -100},          // another year
	}
	got, err := ComputeTotals(decisions)
	if err != nil {
		t.Fatal(err)
	}
	wantMonth := map[string]int64{"2026-04": 1250, "2026-05": 800, "2025-04": 100}
	wantCategory := map[string]int64{"Dining and Drinks": 1500, "Groceries": 350, "Income and Refunds": 300}
	if got.Total != 2150 {
		t.Errorf("total: got %d want 2150", got.Total)
	}
	for k, v := range wantMonth {
		if got.ByMonth[k] != v {
			t.Errorf("month %s: got %d want %d", k, got.ByMonth[k], v)
		}
	}
	for k, v := range wantCategory {
		if got.ByCategory[k] != v {
			t.Errorf("category %s: got %d want %d", k, got.ByCategory[k], v)
		}
	}
	if len(got.ByMonth) != 3 || len(got.ByCategory) != 3 {
		t.Errorf("unexpected extra entries: %+v", got)
	}
}

func TestExport(t *testing.T) {
	files, err := Export([]ClassifiedTransaction{
		classifiedFor("05/02/26", "LATER", -500, "Groceries", 0.9, 0.8),
		classifiedFor("04/22/26", "EARLIER", -211, "Dining and Drinks", 0.9, 0.1),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(files.StatesCSV, "EARLIER") || !strings.Contains(files.StatesJSON, `"key"`) {
		t.Errorf("states missing: %+v", files)
	}
	for _, want := range []string{`"2026-04": 2.11`, `"2026-05": 5.00`, `"total_expenditure": 7.11`} {
		if !strings.Contains(files.TotalsJSON, want) {
			t.Errorf("totals missing %q in %s", want, files.TotalsJSON)
		}
	}
	if _, err := Export([]ClassifiedTransaction{classifiedFor("someday", "X", -1, "Other", 1, 1)}); err == nil {
		t.Error("expected a date error")
	}
}
