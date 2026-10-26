// demo_process_test.go

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func testSources() processSources {
	process := `{"id":"2026-06","label":"June 2026","fileName":"statement-2026-06.pdf","pages":7,"pageReadings":[],"sections":[],"rows":[],"summary":{"rowsFound":1,"depositsRemoved":0,"spendingRows":1}}`
	return processSources{
		Artifacts: fstest.MapFS{
			"2026-06/process.json":      {Data: []byte(process)},
			"2026-06/page-3-before.png": {Data: []byte("before")},
			"secret.txt":                {Data: []byte("secret")},
		},
		Statements: fstest.MapFS{"statement-2026-06.pdf": {Data: []byte("%PDF-fake")}},
	}
}

func testTransactions() []demoTransaction {
	return []demoTransaction{
		{Date: "06/03/26", NeedsReview: true},
		{Date: "06/04/26", NeedsReview: false},
		{Date: "07/04/26", NeedsReview: true},
	}
}

func get(handler http.Handler, path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest("GET", path, nil))
	return recorder
}

func testHandler() http.Handler {
	mux := http.NewServeMux()
	registerProcessRoutes(mux, testSources(), testTransactions())
	return mux
}

func TestCountFlagged(t *testing.T) {
	if got := countFlagged(testTransactions(), "2026-06"); got != 1 {
		t.Errorf("got %d, want 1", got)
	}
}

func TestListStatements(t *testing.T) {
	recorder := get(testHandler(), "/api/demo/statements")
	var got []statementListing
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := statementListing{"2026-06", "June 2026", "statement-2026-06.pdf", 9, 7}
	if len(got) != 1 || got[0] != want {
		t.Errorf("got %+v, want [%+v]", got, want)
	}
}

func TestProcessAddsFlaggedForReview(t *testing.T) {
	recorder := get(testHandler(), "/api/demo/statements/2026-06/process")
	var got statementProcess
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Summary.FlaggedForReview == nil || *got.Summary.FlaggedForReview != 1 {
		t.Errorf("summary: %+v", got.Summary)
	}
}

func TestFilesAndImages(t *testing.T) {
	handler := testHandler()
	pdf := get(handler, "/api/demo/statements/2026-06/file")
	if pdf.Code != 200 || pdf.Header().Get("Content-Type") != "application/pdf" || pdf.Body.String() != "%PDF-fake" {
		t.Errorf("pdf: %d %q", pdf.Code, pdf.Header())
	}
	image := get(handler, "/api/demo/statements/2026-06/pages/3/before.png")
	if image.Code != 200 || image.Header().Get("Content-Type") != "image/png" {
		t.Errorf("image: %d %q", image.Code, image.Header())
	}
}

func TestNotFound(t *testing.T) {
	handler := testHandler()
	for _, path := range []string{
		"/api/demo/statements/2026-07/process",
		"/api/demo/statements/2026-07/file",
		"/api/demo/statements/2026-06/pages/1/before.png",
		"/api/demo/statements/2026-06/pages/3/after.png",
		"/api/demo/statements/2026-06/pages/3/before.gif",
		"/api/demo/statements/2026-06/pages/03/before.png",
		"/api/demo/statements/%2e%2e/process",
		"/api/demo/statements/%2e%2e/file",
		"/api/demo/statements/..%5c..%5csecret.txt/file",
		"/api/demo/statements/2026-06/pages/%2e%2e/before.png",
		"/api/demo/statements/2026-06/pages/3/..%2f..%2fsecret.txt",
	} {
		if got := get(handler, path).Code; got != 404 {
			t.Errorf("%s: got %d, want 404", path, got)
		}
	}
}

// TestEmbeddedArtifacts serves the committed artifacts: four statements, three read pages each.
func TestEmbeddedArtifacts(t *testing.T) {
	mux := http.NewServeMux()
	(&DemoData{}).processRoutes(mux)
	var listing []statementListing
	if err := json.Unmarshal(get(mux, "/api/demo/statements").Body.Bytes(), &listing); err != nil {
		t.Fatal(err)
	}
	if len(listing) != 4 {
		t.Fatalf("got %d statements, want 4", len(listing))
	}
	for _, s := range listing {
		var process statementProcess
		if err := json.Unmarshal(get(mux, "/api/demo/statements/"+s.ID+"/process").Body.Bytes(), &process); err != nil {
			t.Fatal(err)
		}
		if len(process.PageReadings) != 7 || process.Summary.SpendingRows != 50 {
			t.Errorf("%s: %d pages, %d spending rows", s.ID, len(process.PageReadings), process.Summary.SpendingRows)
		}
		if got := get(mux, "/api/demo/statements/"+s.ID+"/pages/4/after.png").Code; got != 200 {
			t.Errorf("%s: page 4 after image: %d", s.ID, got)
		}
	}
}
