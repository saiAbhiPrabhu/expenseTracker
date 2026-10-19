// classify.go
//
// Responsibility: ask JEV the fixed questions about each transaction and return validated results.

package main

import (
	"context"
	"fmt"
)

// QuestionSet holds the fixed questions: Request is what is sent, Choices is what answers are validated against.
type QuestionSet struct {
	Request map[string]any
	Choices map[string]choiceQuestion
}

func BuildQuestionSet(categoriesJSON []byte) (QuestionSet, error) {
	choices, err := BuildCategoryQuestions(categoriesJSON)
	if err != nil {
		return QuestionSet{}, err
	}
	request := map[string]any{essentialQuestionID: EssentialQuestion()}
	for id, question := range choices {
		request[id] = question
	}
	return QuestionSet{Request: request, Choices: choices}, nil
}

// ClassifyTransactions sends one request per transaction. On failure it returns the results
// gathered so far together with the error, so completed (paid-for) calls are not lost.
func ClassifyTransactions(ctx context.Context, client JEVClient, transactions []Transaction, questions QuestionSet) ([]ClassifiedTransaction, error) {
	classified := make([]ClassifiedTransaction, 0, len(transactions))
	for i, transaction := range transactions {
		state, err := toJEVState(transaction)
		if err != nil {
			return classified, fmt.Errorf("transaction %d: %w", i, err)
		}
		response, err := client.Evaluate(ctx, state, questions.Request)
		if err != nil {
			return classified, fmt.Errorf("transaction %d: %w", i, err)
		}
		result, err := ParseClassification(transaction, response, questions.Choices)
		if err != nil {
			return classified, fmt.Errorf("transaction %d: %w", i, err)
		}
		classified = append(classified, result)
	}
	return classified, nil
}
