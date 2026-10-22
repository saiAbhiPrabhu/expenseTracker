package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestWebHandler(t *testing.T) {
	handler := NewWebHandler(fstest.MapFS{
		"index.html":    {Data: []byte("<html>app</html>")},
		"assets/app.js": {Data: []byte("console.log(1)")},
	}, nil)
	cases := []struct {
		path       string
		wantStatus int
		wantBody   string
	}{
		{"/", 200, "app"},
		{"/demo", 200, "app"}, // front-end route falls back to index.html
		{"/login", 200, "app"},
		{"/assets/app.js", 200, "console.log"},
		{"/assets/missing.js", 404, ""}, // a missing asset is a real 404, not the app
		{"/api/health", 200, `"status":"ok"`},
		{"/api/unknown", 404, ""},
	}
	for _, c := range cases {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, c.path, nil))
		if recorder.Code != c.wantStatus || !strings.Contains(recorder.Body.String(), c.wantBody) {
			t.Errorf("%s: got %d %q, want %d containing %q", c.path, recorder.Code, recorder.Body.String(), c.wantStatus, c.wantBody)
		}
	}
}
