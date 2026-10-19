package main

import (
	"encoding/json"
	"testing"
)

var testQuestion = choiceQuestion{Criteria: map[string]string{"A": "a", "B": "b"}}

func TestParseChoiceAnswer(t *testing.T) {
	valid := `{"type":"choice","choice":"A","probabilities":{"A":0.8,"B":0.2},"confidence":0.7}`
	if _, err := ParseChoiceAnswer([]byte(valid), testQuestion); err != nil {
		t.Errorf("valid answer rejected: %v", err)
	}
	rounded := `{"type":"choice","choice":"A","probabilities":{"A":0.8,"B":0.19},"confidence":0.7}`
	if _, err := ParseChoiceAnswer([]byte(rounded), testQuestion); err != nil {
		t.Errorf("answer summing to 0.99 rejected: %v", err)
	}
	invalid := map[string]string{
		"wrong type":      `{"type":"noul","choice":"A","probabilities":{"A":1},"confidence":1}`,
		"unknown choice":  `{"type":"choice","choice":"Z","probabilities":{"A":1},"confidence":1}`,
		"unknown option":  `{"type":"choice","choice":"A","probabilities":{"A":0.5,"Z":0.5},"confidence":1}`,
		"bad sum":         `{"type":"choice","choice":"A","probabilities":{"A":0.5,"B":0.2},"confidence":1}`,
		"bad confidence":  `{"type":"choice","choice":"A","probabilities":{"A":1},"confidence":1.5}`,
		"probability > 1": `{"type":"choice","choice":"A","probabilities":{"A":1.5,"B":-0.5},"confidence":1}`,
	}
	for name, data := range invalid {
		if _, err := ParseChoiceAnswer([]byte(data), testQuestion); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestParseNoulAnswer(t *testing.T) {
	if answer, err := ParseNoulAnswer([]byte(`{"type":"noul","noul":0.93}`)); err != nil || answer.Noul != 0.93 {
		t.Errorf("got %+v, %v", answer, err)
	}
	for _, data := range []string{`{"type":"noul","noul":1.2}`, `{"type":"choice","noul":0.5}`} {
		if _, err := ParseNoulAnswer([]byte(data)); err == nil {
			t.Errorf("%s: expected an error", data)
		}
	}
}

func TestParseClassification(t *testing.T) {
	questions := map[string]choiceQuestion{category20QuestionID: testQuestion, category12QuestionID: testQuestion}
	choice := `{"type":"choice","choice":"B","probabilities":{"A":0.1,"B":0.9},"confidence":0.8}`
	response := jevResponse{Answers: map[string]json.RawMessage{
		category20QuestionID: json.RawMessage(choice),
		category12QuestionID: json.RawMessage(choice),
		essentialQuestionID:  json.RawMessage(`{"type":"noul","noul":0.1}`),
	}}
	got, err := ParseClassification(Transaction{Description: "x"}, response, questions)
	if err != nil || got.Category20.Choice != "B" || got.Essential.Noul != 0.1 {
		t.Fatalf("got %+v, %v", got, err)
	}
	delete(response.Answers, essentialQuestionID)
	if _, err := ParseClassification(Transaction{}, response, questions); err == nil {
		t.Error("expected an error for a missing answer")
	}
}
