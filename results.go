// results.go
//
// Responsibility: turn raw JEV answers into typed, validated results. Each model is built only
// through its Parse function, so a value of these types is known to be valid (pydantic-style:
// declared shape + constraints, rejected at the boundary).

package main

import (
	"encoding/json"
	"fmt"
	"math"
)

const (
	essentialQuestionID = "essential_vs_non_essential"
	// JEV rounds each probability to 2 decimals, so with 13-21 options the total can drift by several hundredths.
	probabilityEpsilon = 0.1
)

// ChoiceAnswer: Choice is one of the question's options and Probabilities is a distribution over them.
type ChoiceAnswer struct {
	Type          string             `json:"type"`
	Choice        string             `json:"choice"`
	Probabilities map[string]float64 `json:"probabilities"`
	Confidence    float64            `json:"confidence"`
}

// NoulAnswer: Noul is the probability the answer is yes, in [0, 1].
type NoulAnswer struct {
	Type string  `json:"type"`
	Noul float64 `json:"noul"`
}

// ClassifiedTransaction is a transaction with JEV's answers to the fixed questions.
type ClassifiedTransaction struct {
	Transaction Transaction
	Category20  ChoiceAnswer
	Category12  ChoiceAnswer
	Essential   NoulAnswer
}

// ParseClassification builds a ClassifiedTransaction from a response, validating every answer
// against the questions that were sent.
func ParseClassification(transaction Transaction, response jevResponse, questions map[string]choiceQuestion) (ClassifiedTransaction, error) {
	category20, err := parseChoiceFrom(response, category20QuestionID, questions)
	if err != nil {
		return ClassifiedTransaction{}, err
	}
	category12, err := parseChoiceFrom(response, category12QuestionID, questions)
	if err != nil {
		return ClassifiedTransaction{}, err
	}
	essentialRaw, ok := response.Answers[essentialQuestionID]
	if !ok {
		return ClassifiedTransaction{}, fmt.Errorf("response has no answer for %q", essentialQuestionID)
	}
	essential, err := ParseNoulAnswer(essentialRaw)
	if err != nil {
		return ClassifiedTransaction{}, fmt.Errorf("%s: %w", essentialQuestionID, err)
	}
	return ClassifiedTransaction{Transaction: transaction, Category20: category20, Category12: category12, Essential: essential}, nil
}

func parseChoiceFrom(response jevResponse, id string, questions map[string]choiceQuestion) (ChoiceAnswer, error) {
	raw, ok := response.Answers[id]
	if !ok {
		return ChoiceAnswer{}, fmt.Errorf("response has no answer for %q", id)
	}
	answer, err := ParseChoiceAnswer(raw, questions[id])
	if err != nil {
		return ChoiceAnswer{}, fmt.Errorf("%s: %w", id, err)
	}
	return answer, nil
}

func ParseChoiceAnswer(data []byte, question choiceQuestion) (ChoiceAnswer, error) {
	var answer ChoiceAnswer
	if err := json.Unmarshal(data, &answer); err != nil {
		return ChoiceAnswer{}, err
	}
	if answer.Type != "choice" {
		return ChoiceAnswer{}, fmt.Errorf("answer type %q, want choice", answer.Type)
	}
	if _, ok := question.Criteria[answer.Choice]; !ok {
		return ChoiceAnswer{}, fmt.Errorf("choice %q is not one of the question's options", answer.Choice)
	}
	if err := validateProbabilities(answer.Probabilities, question); err != nil {
		return ChoiceAnswer{}, err
	}
	if !inUnitRange(answer.Confidence) {
		return ChoiceAnswer{}, fmt.Errorf("confidence %v is outside [0, 1]", answer.Confidence)
	}
	return answer, nil
}

func validateProbabilities(probabilities map[string]float64, question choiceQuestion) error {
	sum := 0.0
	for option, probability := range probabilities {
		if _, ok := question.Criteria[option]; !ok {
			return fmt.Errorf("probability for unknown option %q", option)
		}
		if !inUnitRange(probability) {
			return fmt.Errorf("probability %v for %q is outside [0, 1]", probability, option)
		}
		sum += probability
	}
	if math.Abs(sum-1) > probabilityEpsilon {
		return fmt.Errorf("probabilities sum to %v, want 1", sum)
	}
	return nil
}

func ParseNoulAnswer(data []byte) (NoulAnswer, error) {
	var answer NoulAnswer
	if err := json.Unmarshal(data, &answer); err != nil {
		return NoulAnswer{}, err
	}
	if answer.Type != "noul" {
		return NoulAnswer{}, fmt.Errorf("answer type %q, want noul", answer.Type)
	}
	if !inUnitRange(answer.Noul) {
		return NoulAnswer{}, fmt.Errorf("noul %v is outside [0, 1]", answer.Noul)
	}
	return answer, nil
}

func inUnitRange(value float64) bool {
	return value >= 0 && value <= 1
}
