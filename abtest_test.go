package main

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func classified(description string, choice20, choice12 string) ClassifiedTransaction {
	return ClassifiedTransaction{
		Transaction: Transaction{Date: "01/02/25", Description: description, AmountCents: -500},
		Category20:  ChoiceAnswer{Choice: choice20, Confidence: 0.8},
		Category12:  ChoiceAnswer{Choice: choice12, Confidence: 0.6},
	}
}

func TestCompareCategorySets(t *testing.T) {
	report := CompareCategorySets([]ClassifiedTransaction{
		classified("a", "Groceries", "Groceries"),
		classified("b", "Travel", "Shopping"),
		classified("c", "Fees and Insurance", "Housing and Utilities"),
	})
	if report.Total != 3 || report.Agreements != 1 || report.Unmapped != 1 {
		t.Errorf("got %+v", report)
	}
	if math.Abs(report.AvgConfidence20-0.8) > 1e-9 || report.Rows[1].Agrees || report.Rows[1].MappedTo12 != "Personal Care, Travel, and Gifts" {
		t.Errorf("got %+v", report)
	}
	text := FormatABReport(report)
	for _, want := range []string{"1 of 2 comparable (1 unmapped)", "| NO |", "| n/a |", "-5.00"} {
		if !strings.Contains(text, want) {
			t.Errorf("report missing %q:\n%s", want, text)
		}
	}
}

func TestEveryMappedCategoryExistsInBothSets(t *testing.T) {
	data := mustReadCategories(t)
	questions, err := BuildCategoryQuestions(data)
	if err != nil {
		t.Fatal(err)
	}
	for from, to := range categoryIn12 {
		if _, ok := questions[category20QuestionID].Criteria[from]; !ok {
			t.Errorf("%q is not a categories_20 option", from)
		}
		if _, ok := questions[category12QuestionID].Criteria[to]; !ok {
			t.Errorf("%q is not a categories_12 option", to)
		}
	}
	if unmapped := len(questions[category20QuestionID].Criteria) - len(categoryIn12); unmapped != 1 {
		t.Errorf("%d categories_20 options are unmapped, want only Fees and Insurance", unmapped)
	}
}

func TestClassifyTransactionsAgainstFakeJEV(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"model":"m","answers":{
			"category_20":{"type":"choice","choice":"Groceries","probabilities":{"Groceries":1},"confidence":0.9},
			"category_12":{"type":"choice","choice":"Groceries","probabilities":{"Groceries":1},"confidence":0.9},
			"essential_vs_non_essential":{"type":"noul","noul":0.8}}}`))
	}))
	defer server.Close()
	questions, err := BuildQuestionSet(mustReadCategories(t))
	if err != nil {
		t.Fatal(err)
	}
	client := NewJEVClient("key")
	client.Endpoint = server.URL
	results, err := ClassifyTransactions(context.Background(), client,
		[]Transaction{{Date: "01/02/25", Description: "STORE", AmountCents: -500}}, questions)
	if err != nil || len(results) != 1 || results[0].Category12.Choice != "Groceries" || results[0].Essential.Noul != 0.8 {
		t.Fatalf("got %+v, %v", results, err)
	}
}
