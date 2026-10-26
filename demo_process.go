// demo_process.go
//
// Responsibility: the saved Upload and Process artifacts for the demo (the sample statements and what each
// stage of reading them produced). Owned by the process work; routes are registered from DemoData.Routes in demo.go.

package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"regexp"
)

// The sample pdfs and the output of BuildDemoArtifacts, compiled into the binary.
//
//go:embed demo/artifacts demo/statements/*.pdf
var demoProcessFiles embed.FS

// processSources holds the files the process routes serve: <id>/process.json and page images under
// Artifacts, statement-<id>.pdf under Statements.
type processSources struct {
	Artifacts  fs.FS
	Statements fs.FS
}

type statementListing struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	FileName  string `json:"fileName"`
	SizeBytes int64  `json:"sizeBytes"`
	Pages     int    `json:"pages"`
}

var (
	statementID = regexp.MustCompile(`^\d{4}-\d{2}$`)
	pageNumber  = regexp.MustCompile(`^[1-9]\d{0,2}$`)
	imageKind   = regexp.MustCompile(`^(before|after)\.png$`)
)

func (d *DemoData) processRoutes(mux *http.ServeMux) {
	artifacts, _ := fs.Sub(demoProcessFiles, "demo/artifacts")
	statements, _ := fs.Sub(demoProcessFiles, "demo/statements")
	registerProcessRoutes(mux, processSources{artifacts, statements}, d.transactions)
}

// registerProcessRoutes registers:
//
//	GET /api/demo/statements                          the sample statements
//	GET /api/demo/statements/{id}/file                the sample pdf
//	GET /api/demo/statements/{id}/process             what reading it produced
//	GET /api/demo/statements/{id}/pages/{n}/{image}   before.png / after.png of a read page
//
// Ids and page numbers are checked against fixed patterns, so nothing outside the sources can be named.
func registerProcessRoutes(mux *http.ServeMux, sources processSources, transactions []demoTransaction) {
	mux.HandleFunc("GET /api/demo/statements", func(w http.ResponseWriter, r *http.Request) {
		listing, err := listStatements(sources)
		respondJSONOrError(w, listing, err)
	})
	mux.HandleFunc("GET /api/demo/statements/{id}/file", func(w http.ResponseWriter, r *http.Request) {
		serveFile(w, sources.Statements, "statement-"+r.PathValue("id")+".pdf", r.PathValue("id"), "application/pdf")
	})
	mux.HandleFunc("GET /api/demo/statements/{id}/process", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !statementID.MatchString(id) {
			http.NotFound(w, r)
			return
		}
		process, err := readProcess(sources.Artifacts, id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		flagged := countFlagged(transactions, id)
		process.Summary.FlaggedForReview = &flagged
		writeJSON(w, http.StatusOK, process)
	})
	mux.HandleFunc("GET /api/demo/statements/{id}/pages/{n}/{image}", func(w http.ResponseWriter, r *http.Request) {
		id, n, image := r.PathValue("id"), r.PathValue("n"), r.PathValue("image")
		if !pageNumber.MatchString(n) || !imageKind.MatchString(image) {
			http.NotFound(w, r)
			return
		}
		name := "page-" + n + "-" + image
		serveFile(w, sources.Artifacts, id+"/"+name, id, "image/png")
	})
}

// serveFile writes one file of the sources, or 404 when the id is malformed or the file is missing.
func serveFile(w http.ResponseWriter, files fs.FS, name, id, contentType string) {
	content, err := fs.ReadFile(files, name)
	if !statementID.MatchString(id) || err != nil {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Write(content)
}

func respondJSONOrError(w http.ResponseWriter, value any, err error) {
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, value)
}

func readProcess(artifacts fs.FS, id string) (statementProcess, error) {
	var process statementProcess
	data, err := fs.ReadFile(artifacts, id+"/process.json")
	if err != nil {
		return process, err
	}
	return process, json.Unmarshal(data, &process)
}

// listStatements describes every statement that has saved artifacts, sorted by id.
func listStatements(sources processSources) ([]statementListing, error) {
	dirs, err := fs.ReadDir(sources.Artifacts, ".")
	if err != nil {
		return nil, err
	}
	listing := []statementListing{}
	for _, dir := range dirs {
		if !dir.IsDir() || !statementID.MatchString(dir.Name()) {
			continue
		}
		process, err := readProcess(sources.Artifacts, dir.Name())
		if err != nil {
			return nil, err
		}
		info, err := fs.Stat(sources.Statements, process.FileName)
		if err != nil {
			return nil, err
		}
		listing = append(listing, statementListing{process.ID, process.Label, process.FileName, info.Size(), process.Pages})
	}
	return listing, nil
}

// countFlagged counts the transactions of the statement's month that need review.
func countFlagged(transactions []demoTransaction, id string) int {
	count := 0
	for _, t := range transactions {
		date, err := parseDate(t.Date)
		if err == nil && t.NeedsReview && date.Format("2006-01") == id {
			count++
		}
	}
	return count
}
