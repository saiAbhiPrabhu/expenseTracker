// simplify.go
//
// Responsibility: turn a scrubbed statement title into a short, readable one by removing filler
// words. Everything that is not filler passes through unchanged (cities, truncated or cryptic
// names), so a name is never rewritten, only trimmed.

package main

import (
	"regexp"
	"strings"
)

var (
	// Card-processor prefixes glued to the merchant: "TST*THE BOARD", "SQ *CBG", "DD *DOORDASH".
	processorPrefix = regexp.MustCompile(`(?i)\b(?:TST|SQ|DD)\s*\*\s*`)
	// A token made only of digits and punctuation (dates, store and phone numbers), but not "&".
	numberOrPunctuation = regexp.MustCompile(`^[\d\W_]+$`)
	// Receipt boilerplate that repeats the merchant's website.
	websiteNoise = regexp.MustCompile(`(?i)^(?:amzn\.com/bill\w*|doordash\.com)$`)

	statementKeywords = map[string]bool{"CHECKCARD": true, "PURCHASE": true, "PMNT": true}
	danglingWords     = map[string]bool{"to": true, "from": true}
	stateCodes        = map[string]bool{}
)

func init() {
	for _, code := range strings.Fields("AL AK AZ AR CA CO CT DE DC FL GA HI ID IL IN IA KS KY LA ME MD MA MI MN MS MO MT NE NV NH NJ NM NY NC ND OH OK OR PA RI SC SD TN TX UT VT VA WA WV WI WY") {
		stateCodes[code] = true
	}
}

/*
HEURISTIC CYCLE: how a scrubbed title becomes a readable one

Principle: drop a word only when confident it is filler; everything else passes through. A leftover code
is a small flaw, a deleted piece of a merchant name is not recoverable.

  Input: "CHECKCARD 0717 TST* CEDARS IN UNIVERSI SEATTLE WA [REDACTED]"

  1. STRIP PROCESSOR PREFIXES (vendor vocabulary): "TST*", "SQ *", "DD *" become a space, so
     "TST*THE BOARD" and "TST* THE BOARD" look the same.
  2. SPLIT INTO WORDS on whitespace. Later steps work on words, not characters.
  3. DROP FILLER WORDS, any of:
       - statement keyword (vendor vocabulary): CHECKCARD, PURCHASE, PMNT; "SENT" after PMNT; a leading "SP"
       - contains "[REDACTED]", starts with "#", or ends with "#" or ":" (removed values and their labels)
       - only digits and punctuation, except "&" (dates, phone and store numbers)
       - website boilerplate (vendor vocabulary): "Amzn.com/bill...", "DOORDASH.COM"
     A trailing "*" is trimmed from the rest ("VENMO*" -> "VENMO").  -> "CEDARS IN UNIVERSI SEATTLE WA"
  4. DROP A DANGLING "to" / "from" at the end ("Zelle payment to ..."), unless it is the only word left.
  5. DROP THE TRAILING STATE CODE: the last word only, if it is a US state code. State-like words inside
     the name stay ("IN" in "CEDARS IN UNIVERSI", "CO" in "SEATTLE DUMPLINGS CO").
                                                                  -> "CEDARS IN UNIVERSI SEATTLE"
  6. COLLAPSE IMMEDIATE REPEATS, case-insensitive: "7-ELEVEN 7-ELEVEN" -> "7-ELEVEN",
     "TARGET T-1284 TARGET T-1284" -> "TARGET T-1284". Tries 1-word phrases, then 2-word, until stable.
  7. FALL BACK to the trimmed original title if nothing readable is left.

Why this order
  - Numbers and redacted words go first so the state code really is the last word.
  - The dangling word goes before the state code so each rule sees the true end of the title.
  - Repeats go last: earlier steps remove the words between them.

General vs vendor vocabulary
  - Shape-based, should hold for any bank: the tests in step 3 and steps 4-6.
  - Seeded from the sample statements, extend for other banks: the prefixes, keywords, website
    boilerplate and the US-only state list.

Left alone on purpose: city names, truncated or cryptic names, reference codes after a "*", capitalization.
*/

// Function Contract:
// Input: a scrubbed statement title
// Output: the title without filler words; the trimmed original if nothing readable would remain
// Logic: Computer drops statement keywords, processor prefixes, redacted fields, dates and numbers,
// reference labels, the trailing state code and a dangling "to"/"from", then collapses repeated phrases.
func SimplifyDescription(title string) string {
	tokens := keepContentTokens(strings.Fields(processorPrefix.ReplaceAllString(title, " ")))
	tokens = dropDanglingWord(tokens)
	tokens = dropTrailingStateCode(tokens)
	tokens = collapseRepeats(tokens)
	if len(tokens) == 0 {
		return strings.TrimSpace(title)
	}
	return strings.Join(tokens, " ")
}

func keepContentTokens(tokens []string) []string {
	var kept []string
	previous := ""
	for _, token := range tokens {
		upper := strings.ToUpper(token)
		isFiller := isFillerToken(token) ||
			(upper == "SENT" && previous == "PMNT") || // "PMNT SENT"
			(upper == "SP" && len(kept) == 0) //          Shopify prefix, only at the very start
		previous = upper
		if !isFiller {
			kept = append(kept, strings.TrimRight(token, "*"))
		}
	}
	return kept
}

func isFillerToken(token string) bool {
	return statementKeywords[strings.ToUpper(token)] ||
		strings.Contains(token, "[REDACTED]") ||
		strings.HasPrefix(token, "#") ||
		strings.HasSuffix(token, "#") || strings.HasSuffix(token, ":") ||
		(numberOrPunctuation.MatchString(token) && token != "&") ||
		websiteNoise.MatchString(token)
}

func dropDanglingWord(tokens []string) []string {
	if n := len(tokens); n > 1 && danglingWords[strings.ToLower(tokens[n-1])] {
		return tokens[:n-1]
	}
	return tokens
}

// dropTrailingStateCode removes only the last word, and only when it is a state code, so a
// state-like word inside the name ("CEDARS IN UNIVERSI", "DUMPLINGS CO") is left alone.
func dropTrailingStateCode(tokens []string) []string {
	if n := len(tokens); n > 1 && stateCodes[strings.ToUpper(strings.Trim(tokens[n-1], `)(.,|`))] {
		return tokens[:n-1]
	}
	return tokens
}

// collapseRepeats removes a phrase that is immediately repeated: "A B A B" -> "A B", "X X" -> "X".
func collapseRepeats(tokens []string) []string {
	for changed := true; changed; {
		changed = false
		for size := 1; size <= len(tokens)/2 && !changed; size++ {
			for start := 0; start+2*size <= len(tokens) && !changed; start++ {
				if equalFoldTokens(tokens[start:start+size], tokens[start+size:start+2*size]) {
					tokens = append(tokens[:start+size], tokens[start+2*size:]...)
					changed = true
				}
			}
		}
	}
	return tokens
}

func equalFoldTokens(a, b []string) bool {
	for i := range a {
		if !strings.EqualFold(a[i], b[i]) {
			return false
		}
	}
	return true
}
