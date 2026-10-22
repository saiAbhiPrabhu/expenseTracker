// demo.go
//
// Responsibility: the demo's saved data and its API. The demo replays a cached JEV run on synthetic
// transactions; nothing here calls JEV. The server holds no per-user state: the browser keeps its
// edits and posts them to get a recalculated dashboard.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type demoAnswer struct {
	Choice        string             `json:"choice"`
	Confidence    float64            `json:"confidence"`
	Probabilities map[string]float64 `json:"probabilities"`
}

// demoTransaction is one row as the front end shows it. Category and Essential are the values in
// use (JEV's, until the user changes them).
type demoTransaction struct {
	Key                  string      `json:"key"`
	Date                 string      `json:"date"`
	Title                string      `json:"title"`
	Amount               json.Number `json:"amount"`
	Category20           demoAnswer  `json:"category20"`
	EssentialProbability float64     `json:"essentialProbability"`
	NeedsReview          bool        `json:"needsReview"`
	Category             string      `json:"category"`
	Essential            bool        `json:"essential"`
}

type categoryOption struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// demoEdit is what a user can change on a row; an absent field is left as it was.
type demoEdit struct {
	Category  *string `json:"category"`
	Essential *bool   `json:"essential"`
}

type demoSummary struct {
	TransactionCount int                    `json:"transactionCount"`
	TotalExpenditure json.Number            `json:"totalExpenditure"`
	ByMonth          map[string]json.Number `json:"byMonth"`
	ByCategory       map[string]json.Number `json:"byCategory"`
	Essential        demoEssentialSplit     `json:"essential"`
}

type demoEssentialSplit struct {
	Essential    json.Number `json:"essential"`
	NonEssential json.Number `json:"nonEssential"`
}

// DemoData is the saved demo run.
type DemoData struct {
	transactions  []demoTransaction
	decisions     []Decision // JEV's choices before any edit, in the same order
	indexByKey    map[string]int
	categories    []categoryOption
	validCategory map[string]bool
}

// Function Contract:
// Input: the cached A/B results.json and transaction_categories.json
// Output: the demo data
// Logic: Computer turns each cached answer into a final decision (review flag, essential yes/no),
// and lists the 20 categories plus "Other". Keys must be unique so an edit names exactly one row.
func LoadDemoData(resultsJSON, categoriesJSON []byte) (*DemoData, error) {
	var classified []ClassifiedTransaction
	if err := json.Unmarshal(resultsJSON, &classified); err != nil {
		return nil, fmt.Errorf("demo results: %w", err)
	}
	var sets categorySets
	if err := json.Unmarshal(categoriesJSON, &sets); err != nil {
		return nil, fmt.Errorf("demo categories: %w", err)
	}
	data := &DemoData{indexByKey: map[string]int{}, validCategory: map[string]bool{}}
	for _, category := range sets.Categories20 {
		data.categories = append(data.categories, categoryOption{category.Category, category.Description})
	}
	data.categories = append(data.categories, categoryOption{catchAllCategory, catchAllDescription})
	for _, option := range data.categories {
		data.validCategory[option.Name] = true
	}
	for i, c := range classified {
		decision, err := Decide(c)
		if err != nil {
			return nil, fmt.Errorf("demo transaction %d: %w", i+1, err)
		}
		if _, duplicate := data.indexByKey[decision.Key]; duplicate {
			return nil, fmt.Errorf("demo transaction %d: duplicate key", i+1)
		}
		data.indexByKey[decision.Key] = i
		data.decisions = append(data.decisions, decision)
		data.transactions = append(data.transactions, demoTransaction{
			Key: decision.Key, Date: decision.Date, Title: decision.Title, Amount: formatDollars(decision.AmountCents),
			Category20:           demoAnswer{c.Category20.Choice, c.Category20.Confidence, c.Category20.Probabilities},
			EssentialProbability: decision.EssentialProbability, NeedsReview: decision.NeedsReview,
			Category: decision.Category, Essential: decision.Essential,
		})
	}
	return data, nil
}

// Function Contract:
// Input: edits keyed by transaction key
// Output: the dashboard numbers with the edits applied
// Logic: Computer applies each edit's category and essential flag to JEV's decisions, then totals the
// expenditure (any outflow) per month, per category, overall, and split by essential. Unknown keys or
// categories are errors.
func (d *DemoData) Summarize(edits map[string]demoEdit) (demoSummary, error) {
	decisions := append([]Decision(nil), d.decisions...)
	for key, edit := range edits {
		i, known := d.indexByKey[key]
		if !known {
			return demoSummary{}, fmt.Errorf("unknown transaction key %q", key)
		}
		if edit.Category != nil {
			if !d.validCategory[*edit.Category] {
				return demoSummary{}, fmt.Errorf("unknown category %q", *edit.Category)
			}
			decisions[i].Category = *edit.Category
		}
		if edit.Essential != nil {
			decisions[i].Essential = *edit.Essential
		}
	}
	totals, err := ComputeTotals(decisions)
	if err != nil {
		return demoSummary{}, err
	}
	summary := demoSummary{
		TransactionCount: len(decisions),
		TotalExpenditure: formatDollars(totals.Total),
		ByMonth:          map[string]json.Number{},
		ByCategory:       map[string]json.Number{},
	}
	for month, cents := range totals.ByMonth {
		summary.ByMonth[month] = formatDollars(cents)
	}
	for category, cents := range totals.ByCategory {
		summary.ByCategory[category] = formatDollars(cents)
	}
	var essential, nonEssential int64
	for _, decision := range decisions {
		if decision.AmountCents >= 0 {
			continue
		}
		if decision.Essential {
			essential -= decision.AmountCents
		} else {
			nonEssential -= decision.AmountCents
		}
	}
	summary.Essential = demoEssentialSplit{formatDollars(essential), formatDollars(nonEssential)}
	return summary, nil
}

const maxEditsBodyBytes = 1 << 20

// Routes registers the demo API:
//
//	GET  /api/demo/transactions   every transaction with JEV's answers and the review flag
//	GET  /api/demo/categories     the categories a row can be changed to
//	GET  /api/demo/summary        dashboard numbers for JEV's answers as they are
//	POST /api/demo/summary        the same, with {"edits": {"<key>": {"category": "...", "essential": true}}}
func (d *DemoData) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/demo/transactions", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, d.transactions)
	})
	mux.HandleFunc("GET /api/demo/categories", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, d.categories)
	})
	mux.HandleFunc("GET /api/demo/summary", func(w http.ResponseWriter, r *http.Request) {
		d.respondWithSummary(w, nil)
	})
	mux.HandleFunc("POST /api/demo/summary", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Edits map[string]demoEdit `json:"edits"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxEditsBodyBytes)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body: " + err.Error()})
			return
		}
		d.respondWithSummary(w, body.Edits)
	})
}

func (d *DemoData) respondWithSummary(w http.ResponseWriter, edits map[string]demoEdit) {
	summary, err := d.Summarize(edits)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}
