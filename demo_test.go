package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

const testCategories = `{"categories_20": [
	{"category": "Groceries", "description": "food at home"},
	{"category": "Dining and Drinks", "description": "eating out"}], "categories_12": []}`

func testDemo(t *testing.T) *DemoData {
	t.Helper()
	results := []ClassifiedTransaction{
		classifiedFor("06/02/26", "STORE A", -1000, "Groceries", 0.9, 0.9),       // confident, essential
		classifiedFor("06/10/26", "CAFE B", -500, "Dining and Drinks", 0.3, 0.1), // low confidence: review
		classifiedFor("07/03/26", "CAFE C", -250, "Dining and Drinks", 0.8, 0.2),
		classifiedFor("07/04/26", "REFUND D", 400, "Groceries", 0.9, 0.5), // a credit is not expenditure
	}
	data, err := json.Marshal(results)
	if err != nil {
		t.Fatal(err)
	}
	demo, err := LoadDemoData(data, []byte(testCategories))
	if err != nil {
		t.Fatal(err)
	}
	return demo
}

func TestLoadDemoData(t *testing.T) {
	demo := testDemo(t)
	if len(demo.transactions) != 4 || !demo.transactions[1].NeedsReview || demo.transactions[0].NeedsReview {
		t.Errorf("review flags wrong: %+v", demo.transactions)
	}
	if !demo.transactions[0].Essential || demo.transactions[1].Essential || demo.transactions[0].Category != "Groceries" {
		t.Errorf("decisions wrong: %+v", demo.transactions[0])
	}
	if names := demo.categories; len(names) != 3 || names[2].Name != "Other" {
		t.Errorf("categories should be the 20-set list plus Other: %+v", names)
	}
	dup, _ := json.Marshal([]ClassifiedTransaction{
		classifiedFor("06/02/26", "SAME", -100, "Groceries", 0.9, 0.9),
		classifiedFor("06/02/26", "SAME", -100, "Groceries", 0.9, 0.9),
	})
	if _, err := LoadDemoData(dup, []byte(testCategories)); err == nil {
		t.Error("expected an error for duplicate keys")
	}
}

func TestSummarize(t *testing.T) {
	demo := testDemo(t)
	base, err := demo.Summarize(nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(base.TotalExpenditure) != "17.50" || string(base.ByMonth["2026-06"]) != "15.00" || string(base.ByCategory["Dining and Drinks"]) != "7.50" {
		t.Errorf("baseline: %+v", base)
	}
	if string(base.Essential.Essential) != "10.00" || string(base.Essential.NonEssential) != "7.50" {
		t.Errorf("baseline essential split: %+v", base.Essential)
	}

	cafe, groceries, no := demo.transactions[1].Key, "Groceries", false
	edited, err := demo.Summarize(map[string]demoEdit{cafe: {Category: &groceries, Essential: nil}, demo.transactions[0].Key: {Essential: &no}})
	if err != nil {
		t.Fatal(err)
	}
	if string(edited.ByCategory["Groceries"]) != "15.00" || string(edited.ByCategory["Dining and Drinks"]) != "2.50" {
		t.Errorf("category edit should move the spend: %+v", edited.ByCategory)
	}
	if string(edited.TotalExpenditure) != "17.50" || string(edited.Essential.Essential) != "0.00" || string(edited.Essential.NonEssential) != "17.50" {
		t.Errorf("an edit never changes the total, only how it is split: %+v", edited)
	}
	if again, _ := demo.Summarize(nil); string(again.ByCategory["Groceries"]) != "10.00" {
		t.Error("edits must not leak into the saved data")
	}

	other := "Nonsense"
	if _, err := demo.Summarize(map[string]demoEdit{cafe: {Category: &other}}); err == nil {
		t.Error("expected an error for an unknown category")
	}
	if _, err := demo.Summarize(map[string]demoEdit{"nope": {Essential: &no}}); err == nil {
		t.Error("expected an error for an unknown key")
	}
}

func TestDemoRoutes(t *testing.T) {
	demo := testDemo(t)
	handler := NewWebHandler(fstest.MapFS{"index.html": {Data: []byte("app")}}, demo)
	get := func(path string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		return recorder
	}
	var transactions []demoTransaction
	if rec := get("/api/demo/transactions"); rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &transactions) != nil || len(transactions) != 4 {
		t.Fatalf("transactions: %d %s", rec.Code, rec.Body.String())
	}
	if rec := get("/api/demo/categories"); rec.Code != 200 || !strings.Contains(rec.Body.String(), `"name":"Other"`) {
		t.Errorf("categories: %d %s", rec.Code, rec.Body.String())
	}
	if rec := get("/api/demo/summary"); rec.Code != 200 || !strings.Contains(rec.Body.String(), `"totalExpenditure":17.50`) {
		t.Errorf("summary: %d %s", rec.Code, rec.Body.String())
	}

	post := func(body string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/demo/summary", strings.NewReader(body)))
		return recorder
	}
	edit := `{"edits": {"` + transactions[1].Key + `": {"category": "Groceries"}}}`
	if rec := post(edit); rec.Code != 200 || !strings.Contains(rec.Body.String(), `"Groceries":15.00`) {
		t.Errorf("summary with edit: %d %s", rec.Code, rec.Body.String())
	}
	for name, body := range map[string]string{
		"bad json":         `{`,
		"unknown category": `{"edits": {"` + transactions[1].Key + `": {"category": "Nope"}}}`,
		"unknown key":      `{"edits": {"nope": {"essential": true}}}`,
	} {
		if rec := post(body); rec.Code != 400 || !strings.Contains(rec.Body.String(), `"error"`) {
			t.Errorf("%s: want 400 with an error, got %d %s", name, rec.Code, rec.Body.String())
		}
	}
}

// The saved run in demo/cache must load: unique keys, every answer valid. Skipped when it is absent.
func TestSavedDemoRunLoads(t *testing.T) {
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
	summary, err := demo.Summarize(nil)
	if err != nil || summary.TransactionCount != len(demo.transactions) || len(summary.ByMonth) != 4 {
		t.Errorf("saved run: %+v, %v", summary, err)
	}
	for _, transaction := range demo.transactions {
		if !demo.validCategory[transaction.Category] {
			t.Errorf("JEV answered a category that is not in the list: %q", transaction.Category)
		}
	}
}
