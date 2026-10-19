package main

import "testing"

func TestApplyOCRFixes(t *testing.T) {
	got, err := ApplyOCRFixes("a -177.4) b", []OCRFix{{Old: "-177.4)", New: "-177.41"}})
	if err != nil || got != "a -177.41 b" {
		t.Errorf("got %q, %v", got, err)
	}
	if _, err := ApplyOCRFixes("a b", []OCRFix{{Old: "zzz", New: "y"}}); err == nil {
		t.Error("expected an error when the text to fix is missing")
	}
	if _, err := ApplyOCRFixes("x x", []OCRFix{{Old: "x", New: "y"}}); err == nil {
		t.Error("expected an error when the text to fix is ambiguous")
	}
}
