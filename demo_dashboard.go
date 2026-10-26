// demo_dashboard.go
//
// Responsibility: the dashboard page's numbers (see web/DESIGN.md). The server holds no state: the
// browser posts its edits and gets every table back, so all tables add up to the same total.
// Owned by the dashboard work; routes are registered from DemoData.Routes in demo.go.
//
//	POST /api/demo/dashboard
//	  request  {"edits": {"<key>": {"category": "...", "essential": true}}, "includeMoneyMovement": true}
//	  response (dollars, camelCase; "expenditure" is money that left the account, i.e. any negative amount)
//	    includeMoneyMovement  echo of the request
//	    totalExpenditure      everything below adds up to this
//	    transactionCount      how many outflows are counted
//	    moneyMovement         dollars of Transfers and Peer Payments, Savings and Investing, Debt Payments in
//	                          the data, whether or not they are counted
//	    firstDate, lastDate   "YYYY-MM-DD", over all rows in the data
//	    months                [{month "YYYY-MM", expenditure}] oldest first (months with no spending included)
//	    categories            [{name, expenditure, count}] largest first
//	    essential             {essential, nonEssential}
//	    categoryByMonth       [{category, total, byMonth []}] largest first; byMonth lines up with months
//	    merchants             top 10 [{name, total, count}] by total, rows grouped on SimplifyDescription(title)
//	    largest               top 10 [{date "MM/DD/YY", title, amount, category}] by size
//	    recurring             merchants seen in >= 3 distinct months [{name, typicalAmount (median), months, category}]
//	  Unknown keys or categories, or bad JSON: 400 {"error": "..."}.

package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
)

const peerPaymentsCategory = "Transfers and Peer Payments"

// Categories that are money moving rather than spending. The user can leave them out.
var moneyMovementCategories = map[string]bool{
	peerPaymentsCategory:    true,
	"Savings and Investing": true,
	"Debt Payments":         true,
}

const (
	topListSize           = 10
	recurringMinimumMonth = 3
)

type dashMonth struct {
	Month       string      `json:"month"`
	Expenditure json.Number `json:"expenditure"`
}

type dashCategory struct {
	Name        string      `json:"name"`
	Expenditure json.Number `json:"expenditure"`
	Count       int         `json:"count"`
}

type dashCategoryByMonth struct {
	Category string        `json:"category"`
	Total    json.Number   `json:"total"`
	ByMonth  []json.Number `json:"byMonth"`
}

type dashMerchant struct {
	Name  string      `json:"name"`
	Total json.Number `json:"total"`
	Count int         `json:"count"`
}

type dashLargest struct {
	Date     string      `json:"date"`
	Title    string      `json:"title"`
	Amount   json.Number `json:"amount"`
	Category string      `json:"category"`
}

type dashRecurring struct {
	Name          string      `json:"name"`
	TypicalAmount json.Number `json:"typicalAmount"`
	Months        int         `json:"months"`
	Category      string      `json:"category"`
}

type demoDashboard struct {
	IncludeMoneyMovement bool                  `json:"includeMoneyMovement"`
	TotalExpenditure     json.Number           `json:"totalExpenditure"`
	TransactionCount     int                   `json:"transactionCount"`
	MoneyMovement        json.Number           `json:"moneyMovement"`
	FirstDate            string                `json:"firstDate"`
	LastDate             string                `json:"lastDate"`
	Months               []dashMonth           `json:"months"`
	Categories           []dashCategory        `json:"categories"`
	Essential            demoEssentialSplit    `json:"essential"`
	CategoryByMonth      []dashCategoryByMonth `json:"categoryByMonth"`
	Merchants            []dashMerchant        `json:"merchants"`
	Largest              []dashLargest         `json:"largest"`
	Recurring            []dashRecurring       `json:"recurring"`
}

type merchantGroup struct {
	names      map[string]int // spellings seen, by count
	total      int64
	count      int
	months     map[string]bool
	amounts    []int64
	categories map[string]int64 // dollars (cents) per category
}

// Function Contract:
// Input: edits keyed by transaction key, and whether money movement is counted
// Output: every table of the dashboard with the edits applied
// Logic: Computer applies the edits to JEV's decisions, drops money-movement outflows when they are not
// counted, then totals the rest per month, category, essential flag and merchant (rows grouped on the
// simplified title, ignoring case). Recurring merchants appear in at least 3 distinct months.
func (d *DemoData) Dashboard(edits map[string]demoEdit, includeMoneyMovement bool) (demoDashboard, error) {
	decisions, err := d.applyEdits(edits)
	if err != nil {
		return demoDashboard{}, err
	}

	dash := demoDashboard{IncludeMoneyMovement: includeMoneyMovement}
	var moneyMovement, essential, nonEssential int64
	var counted []Decision
	firstDate, lastDate := "", ""
	allMonths := map[string]bool{}
	for _, decision := range decisions {
		date, err := parseDate(decision.Date)
		if err != nil {
			return demoDashboard{}, err
		}
		iso := date.Format("2006-01-02")
		if firstDate == "" || iso < firstDate {
			firstDate = iso
		}
		if iso > lastDate {
			lastDate = iso
		}
		allMonths[date.Format("2006-01")] = true
		if decision.AmountCents >= 0 {
			continue
		}
		if moneyMovementCategories[decision.Category] {
			moneyMovement -= decision.AmountCents
			if !includeMoneyMovement {
				continue
			}
		}
		counted = append(counted, decision)
		if decision.Essential {
			essential -= decision.AmountCents
		} else {
			nonEssential -= decision.AmountCents
		}
	}
	totals, err := ComputeTotals(counted)
	if err != nil {
		return demoDashboard{}, err
	}
	dash.TotalExpenditure = formatDollars(totals.Total)
	dash.TransactionCount = len(counted)
	dash.MoneyMovement = formatDollars(moneyMovement)
	dash.FirstDate, dash.LastDate = firstDate, lastDate
	dash.Essential = demoEssentialSplit{formatDollars(essential), formatDollars(nonEssential)}

	monthKeys := make([]string, 0, len(allMonths))
	for month := range allMonths {
		monthKeys = append(monthKeys, month)
	}
	sort.Strings(monthKeys)
	dash.Months = []dashMonth{}
	for _, month := range monthKeys {
		dash.Months = append(dash.Months, dashMonth{month, formatDollars(totals.ByMonth[month])})
	}

	counts := map[string]int{}
	byCategoryMonth := map[string]map[string]int64{}
	groups := map[string]*merchantGroup{}
	for _, decision := range counted {
		date, _ := parseDate(decision.Date) // already parsed above
		month := date.Format("2006-01")
		spend := -decision.AmountCents
		counts[decision.Category]++
		if byCategoryMonth[decision.Category] == nil {
			byCategoryMonth[decision.Category] = map[string]int64{}
		}
		byCategoryMonth[decision.Category][month] += spend

		if decision.Category == peerPaymentsCategory {
			continue // recipients are anonymized, so different people look like one "merchant"
		}
		name := SimplifyDescription(decision.Title)
		id := strings.ToLower(name)
		group := groups[id]
		if group == nil {
			group = &merchantGroup{names: map[string]int{}, months: map[string]bool{}, categories: map[string]int64{}}
			groups[id] = group
		}
		group.names[name]++
		group.total += spend
		group.count++
		group.months[month] = true
		group.amounts = append(group.amounts, spend)
		group.categories[decision.Category] += spend
	}

	categoryNames := make([]string, 0, len(totals.ByCategory))
	for name := range totals.ByCategory {
		categoryNames = append(categoryNames, name)
	}
	sort.Slice(categoryNames, func(i, j int) bool {
		a, b := categoryNames[i], categoryNames[j]
		if totals.ByCategory[a] != totals.ByCategory[b] {
			return totals.ByCategory[a] > totals.ByCategory[b]
		}
		return a < b
	})
	dash.Categories = []dashCategory{}
	dash.CategoryByMonth = []dashCategoryByMonth{}
	for _, name := range categoryNames {
		dash.Categories = append(dash.Categories, dashCategory{name, formatDollars(totals.ByCategory[name]), counts[name]})
		row := dashCategoryByMonth{Category: name, Total: formatDollars(totals.ByCategory[name])}
		for _, month := range monthKeys {
			row.ByMonth = append(row.ByMonth, formatDollars(byCategoryMonth[name][month]))
		}
		dash.CategoryByMonth = append(dash.CategoryByMonth, row)
	}

	type namedGroup struct {
		name  string
		group *merchantGroup
	}
	var merchants []namedGroup
	for _, group := range groups {
		merchants = append(merchants, namedGroup{mostCommon(group.names, nil), group})
	}
	sort.Slice(merchants, func(i, j int) bool {
		a, b := merchants[i], merchants[j]
		if a.group.total != b.group.total {
			return a.group.total > b.group.total
		}
		return a.name < b.name
	})
	dash.Merchants = []dashMerchant{}
	for i, m := range merchants {
		if i == topListSize {
			break
		}
		dash.Merchants = append(dash.Merchants, dashMerchant{m.name, formatDollars(m.group.total), m.group.count})
	}
	type typical struct {
		row   dashRecurring
		cents int64
	}
	var recurring []typical
	for _, m := range merchants {
		if len(m.group.months) < recurringMinimumMonth {
			continue
		}
		cents := median(m.group.amounts)
		recurring = append(recurring, typical{dashRecurring{m.name, formatDollars(cents), len(m.group.months),
			mostCommon(nil, m.group.categories)}, cents})
	}
	sort.SliceStable(recurring, func(i, j int) bool { return recurring[i].cents > recurring[j].cents })
	dash.Recurring = []dashRecurring{}
	for _, r := range recurring {
		dash.Recurring = append(dash.Recurring, r.row)
	}

	sorted := append([]Decision(nil), counted...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].AmountCents != sorted[j].AmountCents {
			return sorted[i].AmountCents < sorted[j].AmountCents
		}
		return sorted[i].Title < sorted[j].Title
	})
	dash.Largest = []dashLargest{}
	for i, decision := range sorted {
		if i == topListSize {
			break
		}
		dash.Largest = append(dash.Largest, dashLargest{decision.Date, decision.Title, formatDollars(-decision.AmountCents), decision.Category})
	}
	return dash, nil
}

// mostCommon returns the key with the highest count (from counts, or from cents when counts is nil),
// the smallest name on ties.
func mostCommon(counts map[string]int, cents map[string]int64) string {
	best, bestValue := "", int64(-1)
	consider := func(name string, value int64) {
		if value > bestValue || (value == bestValue && name < best) {
			best, bestValue = name, value
		}
	}
	for name, n := range counts {
		consider(name, int64(n))
	}
	for name, n := range cents {
		consider(name, n)
	}
	return best
}

// median of cents; for an even count, the average of the two middle values rounded to a cent.
func median(values []int64) int64 {
	sorted := append([]int64(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	n := len(sorted)
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2] + 1) / 2
}

func (d *DemoData) dashboardRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/demo/dashboard", func(w http.ResponseWriter, r *http.Request) {
		body := struct {
			Edits                map[string]demoEdit `json:"edits"`
			IncludeMoneyMovement bool                `json:"includeMoneyMovement"`
		}{IncludeMoneyMovement: true}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxEditsBodyBytes)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body: " + err.Error()})
			return
		}
		dash, err := d.Dashboard(body.Edits, body.IncludeMoneyMovement)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, dash)
	})
}
