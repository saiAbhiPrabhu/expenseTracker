package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const dashCategories = `{"categories_20": [
	{"category": "Groceries", "description": "food at home"},
	{"category": "Dining and Drinks", "description": "eating out"},
	{"category": "Housing and Utilities", "description": "rent"},
	{"category": "Transfers and Peer Payments", "description": "moves"}], "categories_12": []}`

func dashDemo(t *testing.T) *DemoData {
	t.Helper()
	results := []ClassifiedTransaction{
		classifiedFor("06/01/26", "RENT PAYMENT", -100000, "Housing and Utilities", 0.9, 0.9),
		classifiedFor("07/01/26", "Rent Payment", -100000, "Housing and Utilities", 0.9, 0.9),
		classifiedFor("08/01/26", "RENT PAYMENT", -100200, "Housing and Utilities", 0.9, 0.9),
		classifiedFor("06/05/26", "CAFE B", -500, "Dining and Drinks", 0.9, 0.1),
		classifiedFor("07/05/26", "CAFE B", -700, "Dining and Drinks", 0.9, 0.1),
		classifiedFor("07/09/26", "STORE A", -3000, "Groceries", 0.9, 0.9),
		classifiedFor("08/09/26", "ZELLE TO FRIEND", -20000, "Transfers and Peer Payments", 0.9, 0.1),
		classifiedFor("08/10/26", "PAYCHECK", 500000, "Groceries", 0.9, 0.5), // a credit is not expenditure
	}
	data, _ := json.Marshal(results)
	demo, err := LoadDemoData(data, []byte(dashCategories))
	if err != nil {
		t.Fatal(err)
	}
	return demo
}

func cents(n json.Number) int64 {
	f, err := n.Float64()
	if err != nil {
		panic(err)
	}
	return int64(f*100 + map[bool]float64{true: 0.5, false: -0.5}[f >= 0])
}

// checkAddsUp asserts every table matches the overview total.
func checkAddsUp(t *testing.T, dash demoDashboard) {
	t.Helper()
	total := cents(dash.TotalExpenditure)
	var sum int64
	for _, m := range dash.Months {
		sum += cents(m.Expenditure)
	}
	if sum != total {
		t.Errorf("months sum %d, total %d", sum, total)
	}
	sum = 0
	count := 0
	for _, c := range dash.Categories {
		sum += cents(c.Expenditure)
		count += c.Count
	}
	if sum != total || count != dash.TransactionCount {
		t.Errorf("categories sum %d count %d, total %d count %d", sum, count, total, dash.TransactionCount)
	}
	if cents(dash.Essential.Essential)+cents(dash.Essential.NonEssential) != total {
		t.Errorf("essential split does not add up")
	}
	var grid int64
	for _, row := range dash.CategoryByMonth {
		var rowSum int64
		for _, v := range row.ByMonth {
			rowSum += cents(v)
		}
		if rowSum != cents(row.Total) || len(row.ByMonth) != len(dash.Months) {
			t.Errorf("category by month row %q: %d vs %d", row.Category, rowSum, cents(row.Total))
		}
		grid += rowSum
	}
	if grid != total {
		t.Errorf("category by month sums to %d, total %d", grid, total)
	}
}

func TestDashboardTotalsAndToggle(t *testing.T) {
	demo := dashDemo(t)
	on, err := demo.Dashboard(nil, true)
	if err != nil {
		t.Fatal(err)
	}
	off, err := demo.Dashboard(nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if cents(on.TotalExpenditure) != 100000+100000+100200+500+700+3000+20000 || on.TransactionCount != 7 {
		t.Errorf("total %s count %d", on.TotalExpenditure, on.TransactionCount)
	}
	if cents(on.MoneyMovement) != 20000 || cents(off.MoneyMovement) != 20000 {
		t.Errorf("money movement should be reported either way: %s %s", on.MoneyMovement, off.MoneyMovement)
	}
	if cents(on.TotalExpenditure)-cents(off.TotalExpenditure) != 20000 || on.TransactionCount-off.TransactionCount != 1 {
		t.Errorf("toggle should change the total by exactly the money movement")
	}
	checkAddsUp(t, on)
	checkAddsUp(t, off)
	if on.FirstDate != "2026-06-01" || on.LastDate != "2026-08-10" || len(on.Months) != 3 {
		t.Errorf("range %s..%s months %v", on.FirstDate, on.LastDate, on.Months)
	}
	if on.Categories[0].Name != "Housing and Utilities" {
		t.Errorf("categories not largest first: %+v", on.Categories)
	}
	for _, c := range off.Categories {
		if moneyMovementCategories[c.Name] {
			t.Errorf("money movement category still listed when off: %s", c.Name)
		}
	}
	for _, l := range off.Largest {
		if moneyMovementCategories[l.Category] {
			t.Errorf("largest lists money movement when off")
		}
	}
	if on.Largest[0].Amount != "1002.00" || on.Largest[1].Category == "" {
		t.Errorf("largest wrong: %+v", on.Largest)
	}
}

func TestDashboardMerchantsAndRecurring(t *testing.T) {
	dash, err := dashDemo(t).Dashboard(nil, true)
	if err != nil {
		t.Fatal(err)
	}
	// "RENT PAYMENT" and "Rent Payment" are one merchant, whatever the case.
	if len(dash.Merchants) == 0 || dash.Merchants[0].Count != 3 || dash.Merchants[0].Total != "3002.00" {
		t.Errorf("merchants: %+v", dash.Merchants)
	}
	if len(dash.Recurring) != 1 {
		t.Fatalf("only rent is in 3 months: %+v", dash.Recurring)
	}
	r := dash.Recurring[0]
	if r.Months != 3 || r.TypicalAmount != "1000.00" || r.Category != "Housing and Utilities" {
		t.Errorf("recurring: %+v", r)
	}
	if median([]int64{1, 2, 4, 9}) != 3 || median([]int64{5, 1, 3}) != 3 {
		t.Errorf("median wrong")
	}
}

func TestDashboardEditsMoveMoney(t *testing.T) {
	demo := dashDemo(t)
	before, _ := demo.Dashboard(nil, true)
	cafe := demo.transactions[3].Key
	groceries, yes := "Groceries", true
	after, err := demo.Dashboard(map[string]demoEdit{cafe: {Category: &groceries, Essential: &yes}}, true)
	if err != nil {
		t.Fatal(err)
	}
	checkAddsUp(t, after)
	if before.TotalExpenditure != after.TotalExpenditure {
		t.Errorf("an edit must not change the total")
	}
	amount := func(d demoDashboard, name string) int64 {
		for _, c := range d.Categories {
			if c.Name == name {
				return cents(c.Expenditure)
			}
		}
		return 0
	}
	if amount(after, "Groceries")-amount(before, "Groceries") != 500 || amount(before, "Dining and Drinks")-amount(after, "Dining and Drinks") != 500 {
		t.Errorf("the 5.00 should move from Dining to Groceries")
	}
	if cents(after.Essential.Essential)-cents(before.Essential.Essential) != 500 {
		t.Errorf("essential should grow by 5.00")
	}
	// Editing a row into money movement removes it when money movement is off.
	transfers := "Transfers and Peer Payments"
	off, _ := demo.Dashboard(map[string]demoEdit{cafe: {Category: &transfers}}, false)
	base, _ := demo.Dashboard(nil, false)
	if cents(base.TotalExpenditure)-cents(off.TotalExpenditure) != 500 {
		t.Errorf("a row edited into money movement should drop out when it is off")
	}
}

func TestDashboardBadInput(t *testing.T) {
	demo := dashDemo(t)
	if _, err := demo.Dashboard(map[string]demoEdit{"nope": {}}, true); err == nil {
		t.Error("unknown key should fail")
	}
	bad := "Nonsense"
	if _, err := demo.Dashboard(map[string]demoEdit{demo.transactions[0].Key: {Category: &bad}}, true); err == nil {
		t.Error("unknown category should fail")
	}
	mux := http.NewServeMux()
	demo.Routes(mux)
	post := func(body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/api/demo/dashboard", strings.NewReader(body)))
		return w
	}
	for _, body := range []string{`{`, `{"edits":{"nope":{}}}`, `{"edits":{"` + demo.transactions[0].Key + `":{"category":"Nonsense"}}}`} {
		w := post(body)
		var out map[string]string
		json.Unmarshal(w.Body.Bytes(), &out)
		if w.Code != http.StatusBadRequest || out["error"] == "" {
			t.Errorf("%s: want 400 with error, got %d %s", body, w.Code, w.Body)
		}
	}
	// money movement is counted when the field is absent
	w := post(`{}`)
	var dash demoDashboard
	json.Unmarshal(w.Body.Bytes(), &dash)
	if w.Code != 200 || !dash.IncludeMoneyMovement || dash.TransactionCount != 7 {
		t.Errorf("default request: %d %s", w.Code, w.Body)
	}
}

func TestDashboardRealRun(t *testing.T) {
	results, err := os.ReadFile("demo/cache/results.json")
	if err != nil {
		t.Skip("demo/cache/results.json not present")
	}
	categories, err := os.ReadFile("transaction_categories.json")
	if err != nil {
		t.Fatal(err)
	}
	demo, err := LoadDemoData(results, categories)
	if err != nil {
		t.Fatal(err)
	}
	on, err := demo.Dashboard(nil, true)
	if err != nil {
		t.Fatal(err)
	}
	off, _ := demo.Dashboard(nil, false)
	checkAddsUp(t, on)
	checkAddsUp(t, off)
	if cents(on.TotalExpenditure)-cents(off.TotalExpenditure) != cents(on.MoneyMovement) {
		t.Errorf("toggle changed the total by %d, money movement is %d", cents(on.TotalExpenditure)-cents(off.TotalExpenditure), cents(on.MoneyMovement))
	}
	// the dashboard agrees with the existing summary endpoint
	summary, err := demo.Summarize(nil)
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalExpenditure != on.TotalExpenditure {
		t.Errorf("summary %s, dashboard %s", summary.TotalExpenditure, on.TotalExpenditure)
	}
	if len(on.Merchants) != 10 || len(on.Largest) != 10 {
		t.Errorf("want 10 merchants and 10 largest, got %d and %d", len(on.Merchants), len(on.Largest))
	}
	for i := 1; i < len(on.Largest); i++ {
		if cents(on.Largest[i].Amount) > cents(on.Largest[i-1].Amount) {
			t.Errorf("largest not in order")
		}
	}
	for _, r := range on.Recurring {
		if r.Months < 3 {
			t.Errorf("recurring with %d months: %+v", r.Months, r)
		}
	}
	t.Logf("total %s, money movement %s, recurring %d", on.TotalExpenditure, on.MoneyMovement, len(on.Recurring))
	for _, r := range on.Recurring {
		t.Logf("recurring: %+v", r)
	}
}

// Anonymized peer payments count in the totals but are not "merchants": different people look the same.
func TestDashboardPeerPaymentsAreNotMerchants(t *testing.T) {
	results := []ClassifiedTransaction{
		classifiedFor("06/03/26", "Zelle payment to [REDACTED]", -1500, "Transfers and Peer Payments", 0.9, 0.1),
		classifiedFor("07/03/26", "Zelle payment to [REDACTED]", -1500, "Transfers and Peer Payments", 0.9, 0.1),
		classifiedFor("08/03/26", "Zelle payment to [REDACTED]", -1500, "Transfers and Peer Payments", 0.9, 0.1),
		classifiedFor("06/04/26", "STORE A", -700, "Groceries", 0.9, 0.9),
	}
	data, _ := json.Marshal(results)
	demo, err := LoadDemoData(data, []byte(testCategories))
	if err != nil {
		t.Fatal(err)
	}
	dash, err := demo.Dashboard(nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if cents(dash.TotalExpenditure) != 5200 || cents(dash.MoneyMovement) != 4500 {
		t.Errorf("peer payments must still be counted: total %s, money movement %s", dash.TotalExpenditure, dash.MoneyMovement)
	}
	if len(dash.Merchants) != 1 || dash.Merchants[0].Name != "STORE A" || len(dash.Recurring) != 0 {
		t.Errorf("merchants %+v, recurring %+v", dash.Merchants, dash.Recurring)
	}
}
