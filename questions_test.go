package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestBuildCategoryQuestionsAddsCatchAll(t *testing.T) {
	questions, err := BuildCategoryQuestions([]byte(`{
		"categories_20": [{"category": "A", "description": "a"}, {"category": "B", "description": "b"}],
		"categories_12": [{"category": "C", "description": "c"}]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	for id, wantSize := range map[string]int{category20QuestionID: 3, category12QuestionID: 2} {
		question := questions[id]
		if question.Type != "choice" || len(question.Criteria) != wantSize {
			t.Errorf("%s: got %+v, want a choice with %d options", id, question, wantSize)
		}
		if question.Criteria[catchAllCategory] != catchAllDescription {
			t.Errorf("%s: missing catch-all option", id)
		}
	}
}

func TestBuildCategoryQuestionsRealFile(t *testing.T) {
	data, err := os.ReadFile("transaction_categories.json")
	if err != nil {
		t.Fatal(err)
	}
	questions, err := BuildCategoryQuestions(data)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(questions[category20QuestionID].Criteria); got != 21 {
		t.Errorf("category_20 has %d options, want 21", got)
	}
	if got := len(questions[category12QuestionID].Criteria); got != 13 {
		t.Errorf("category_12 has %d options, want 13", got)
	}
}

func TestEssentialQuestionJSON(t *testing.T) {
	got, err := json.Marshal(EssentialQuestion())
	if err != nil {
		t.Fatal(err)
	}
	want := `{"type":"noul","instructions":"Is the transaction an essential expense?",` +
		`"criteria":{"true":"Essential expenses are needed for survival, work, or school.",` +
		`"false":"Non-essential expenses are optional and can be cut without real hardship."}}`
	if string(got) != want {
		t.Errorf("got %s\nwant %s", got, want)
	}
}
