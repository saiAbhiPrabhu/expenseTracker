package main

import "testing"

func TestSimplifyDescription(t *testing.T) {
	cases := []struct{ title, want string }{
		{"CHECKCARD 0421 DUNKIN #[REDACTED] COLLEGE PARK MD [REDACTED] [REDACTED]", "DUNKIN COLLEGE PARK"},
		{"CHECKCARD 0511 TST*THE BOARD AND BREW College Park MD [REDACTED]", "THE BOARD AND BREW College Park"},
		{"CHECKCARD 0627 TST* PASTA & CO. - UNIV SEATTLE WA [REDACTED]", "PASTA & CO. UNIV SEATTLE"},
		{"CHECKCARD 0717 TST* CEDARS IN UNIVERSI SEATTLE WA [REDACTED]", "CEDARS IN UNIVERSI SEATTLE"}, // IN is not the state
		{"CHECKCARD 0704 SEATTLE DUMPLINGS CO 131-[REDACTED] IL [REDACTED]", "SEATTLE DUMPLINGS CO"},   // only the last code goes
		{"CHECKCARD 0710 REI #11 SEATTLE SEATTLE WA [REDACTED]", "REI SEATTLE"},
		{"TARGET T-1284 06/28 #[REDACTED] PURCHASE TARGET T-1284 Seattle WA", "TARGET T-1284 Seattle"},
		{"7-ELEVEN 08/29 #[REDACTED] PURCHASE 7-ELEVEN College Park MD", "7-ELEVEN College Park"},
		{"Zelle payment to [REDACTED]", "Zelle payment"},
		{"PMNT SENT 0620 VENMO* [REDACTED]", "VENMO"},
		{"PURCHASE 0821 VENMO *[REDACTED]", "VENMO"},
		{"Online Banking transfer to CHK 0671 Confirmation# [REDACTED]", "Online Banking transfer to CHK"},
		{"PURCHASE 0729 DD *DOORDASH RAMENDANB DOORDASH.COM CA", "DOORDASH RAMENDANB"},
		{"PURCHASE 0630 Amazon.com*036XC55V3 Amzn.com/billWA", "Amazon.com*036XC55V3"},
		{"Article UW JV II DES:[REDACTED] ID:[REDACTED] INDN:[REDACTED] ID:[REDACTED]", "Article UW JV II"},
		{"CHECKCARD 0629 SP TABLETOP VILLAGE 120-[REDACTED] WA [REDACTED] 7", "TABLETOP VILLAGE"},
		{"CHECKCARD 0511 GT COLLEGE PARK COLLEGE PARK MD [REDACTED]", "GT COLLEGE PARK"},
		{"CHECKCARD 0001 [REDACTED]", "CHECKCARD 0001 [REDACTED]"}, // nothing readable left: keep the original
	}
	for _, c := range cases {
		if got := SimplifyDescription(c.title); got != c.want {
			t.Errorf("%q\n got  %q\n want %q", c.title, got, c.want)
		}
	}
}
