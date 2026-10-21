// decide.go
//
// Responsibility: turn JEV's raw probabilities into the final values the app stores and shows.
// The 20-category set is the chosen set; the 12-category answers are only used by the A/B test.

package main

const (
	// A category answer with confidence below this goes to human review.
	reviewConfidenceThreshold = 0.5
	// An essential probability at or above this counts as essential.
	essentialThreshold = 0.5
)

// Decision is the final, app-facing record for one transaction.
type Decision struct {
	Key                  string
	Date                 string
	Title                string
	AmountCents          int64
	Category             string
	Confidence           float64
	NeedsReview          bool
	Essential            bool
	EssentialProbability float64
}

// Function Contract:
// Input: a classified transaction
// Output: its final Decision
// Logic: Computer takes the 20-set category, flags it for review when its confidence is below the
// threshold, and turns the essential probability into a yes/no.
func Decide(classified ClassifiedTransaction) (Decision, error) {
	key, err := TransactionKey(classified.Transaction)
	if err != nil {
		return Decision{}, err
	}
	return Decision{
		Key:                  key,
		Date:                 classified.Transaction.Date,
		Title:                classified.Transaction.Description,
		AmountCents:          classified.Transaction.AmountCents,
		Category:             classified.Category20.Choice,
		Confidence:           classified.Category20.Confidence,
		NeedsReview:          classified.Category20.Confidence < reviewConfidenceThreshold,
		Essential:            classified.Essential.Noul >= essentialThreshold,
		EssentialProbability: classified.Essential.Noul,
	}, nil
}
