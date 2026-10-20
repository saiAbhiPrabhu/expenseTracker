package main

import "testing"

func TestParseTitlesCSV(t *testing.T) {
	data := "month,row,date,amount,title,note\r\nmay,1,04/22/26,-2.11,\"DUNKIN, COLLEGE PARK\",x\r\njul,2,06/22/26,-1295.95,ZELLE,\r\n"
	got, err := ParseTitlesCSV(data)
	if err != nil || len(got) != 2 {
		t.Fatalf("got %+v, %v", got, err)
	}
	if got[0] != (Transaction{Date: "04/22/26", Description: "DUNKIN, COLLEGE PARK", AmountCents: -211}) {
		t.Errorf("row 1: %+v", got[0])
	}
	if got[1].AmountCents != -129595 {
		t.Errorf("row 2: %+v", got[1])
	}
	for _, bad := range []string{"", "date,title\r\n1,2\r\n", "date,amount,title\r\n01/01/26,12,X\r\n"} {
		if _, err := ParseTitlesCSV(bad); err == nil {
			t.Errorf("expected an error for %q", bad)
		}
	}
}
