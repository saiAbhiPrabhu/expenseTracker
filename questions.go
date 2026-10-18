// questions.go

package main

import "encoding/json"

const (
	category20QuestionID = "category_20"
	category12QuestionID = "category_12"
	catchAllCategory     = "Other"
	catchAllDescription  = "Anything that fits none of the other categories."
)

// choiceQuestion is a JEV "choice" question: pick one option from Criteria.
type choiceQuestion struct {
	Type         string            `json:"type"`
	Instructions string            `json:"instructions"`
	Criteria     map[string]string `json:"criteria"`
}

type categoryEntry struct {
	Category    string `json:"category"`
	Description string `json:"description"`
}

// Used for A/B testing of category sets
type categorySets struct {
	Categories20 []categoryEntry `json:"categories_20"`
	Categories12 []categoryEntry `json:"categories_12"`
}

// Function Contract:
// Input: contents of transaction_categories.json
// Output: the two category questions, keyed by question ID
// Logic: Computer builds one Choice question per category set, each with a catch-all "Other" option added.
func BuildCategoryQuestions(categoriesJSON []byte) (map[string]choiceQuestion, error) {
	var sets categorySets
	if err := json.Unmarshal(categoriesJSON, &sets); err != nil {
		return nil, err
	}
	return map[string]choiceQuestion{
		category20QuestionID: newCategoryQuestion(sets.Categories20),
		category12QuestionID: newCategoryQuestion(sets.Categories12),
	}, nil
}

func newCategoryQuestion(categories []categoryEntry) choiceQuestion {
	criteria := make(map[string]string, len(categories)+1)
	for _, category := range categories {
		criteria[category.Category] = category.Description
	}
	criteria[catchAllCategory] = catchAllDescription
	return choiceQuestion{
		Type:         "choice",
		Instructions: "Which category best fits the transaction?",
		Criteria:     criteria,
	}
}

// noulQuestion is a JEV "noul" question: the answer is the probability of yes.
type noulQuestion struct {
	Type         string       `json:"type"`
	Instructions string       `json:"instructions"`
	Criteria     noulCriteria `json:"criteria"`
}

type noulCriteria struct {
	True  string `json:"true"`
	False string `json:"false"`
}

// EssentialQuestion asks whether the expense is essential; the criteria split one definition into yes and no.
func EssentialQuestion() noulQuestion {
	return noulQuestion{
		Type:         "noul",
		Instructions: "Is the transaction an essential expense?",
		Criteria: noulCriteria{
			True:  "Essential expenses are needed for survival, work, or school.",
			False: "Non-essential expenses are optional and can be cut without real hardship.",
		},
	}
}
