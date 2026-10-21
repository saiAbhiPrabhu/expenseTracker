package main

import "testing"

func classifiedFor(date, title string, cents int64, category string, confidence, essential float64) ClassifiedTransaction {
	return ClassifiedTransaction{
		Transaction: Transaction{Date: date, Description: title, AmountCents: cents},
		Category20:  ChoiceAnswer{Choice: category, Confidence: confidence},
		Essential:   NoulAnswer{Noul: essential},
	}
}

func TestDecide(t *testing.T) {
	cases := []struct {
		confidence, essential     float64
		wantReview, wantEssential bool
	}{
		{0.49, 0.49, true, false},
		{0.50, 0.50, false, true}, // review only below 0.5; essential from 0.5
		{0.93, 0.10, false, false},
	}
	for _, c := range cases {
		got, err := Decide(classifiedFor("04/22/26", "X", -100, "Dining and Drinks", c.confidence, c.essential))
		if err != nil || got.NeedsReview != c.wantReview || got.Essential != c.wantEssential || got.Category != "Dining and Drinks" {
			t.Errorf("confidence %v essential %v: got %+v, %v", c.confidence, c.essential, got, err)
		}
	}
}
