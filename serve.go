// serve.go
//
// Responsibility: the web server. It serves the built front end (web/dist) and the /api routes.
// Pure handler construction: the directory and the listening address are chosen in main.go.

package main

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// NewWebHandler serves the single-page app from site, plus the demo API when demo is not nil. A path
// with no file extension that is not an asset (for example /demo) gets index.html so the front-end
// router can handle it.
func NewWebHandler(site fs.FS, demo *DemoData) http.Handler {
	mux := http.NewServeMux()
	if demo != nil {
		demo.Routes(mux)
	}
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	files := http.FileServerFS(site)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if isPageRoute(r.URL.Path) {
			http.ServeFileFS(w, r, site, "index.html")
			return
		}
		files.ServeHTTP(w, r)
	})
	return mux
}

// isPageRoute is true for the root and for extensionless paths, which are front-end routes.
func isPageRoute(urlPath string) bool {
	return urlPath == "/" || (path.Ext(urlPath) == "" && !strings.HasPrefix(urlPath, "/api/"))
}
