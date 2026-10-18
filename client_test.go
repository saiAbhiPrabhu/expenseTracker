package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testClient(url string, sleeps *[]time.Duration) JEVClient {
	client := NewJEVClient("key")
	client.Endpoint = url
	client.Sleep = func(d time.Duration) { *sleeps = append(*sleeps, d) }
	return client
}

func TestEvaluateRetriesRateLimitWithBackoff(t *testing.T) {
	statuses := []int{429, 529, 200}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer key" {
			t.Errorf("missing bearer token")
		}
		w.WriteHeader(statuses[calls])
		if statuses[calls] == 200 {
			w.Write([]byte(`{"model":"jev-1","answers":{"q":{"type":"noul","noul":0.5}}}`))
		}
		calls++
	}))
	defer server.Close()
	var sleeps []time.Duration
	response, err := testClient(server.URL, &sleeps).Evaluate(context.Background(), "state", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 3 || len(sleeps) != 2 || sleeps[0] != time.Second || sleeps[1] != 2*time.Second {
		t.Errorf("calls=%d sleeps=%v, want 3 calls and sleeps [1s 2s]", calls, sleeps)
	}
	if _, ok := response.Answers["q"]; !ok {
		t.Errorf("answers not parsed: %+v", response)
	}
}

func TestEvaluateDoesNotRetryOtherErrors(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(401)
	}))
	defer server.Close()
	var sleeps []time.Duration
	if _, err := testClient(server.URL, &sleeps).Evaluate(context.Background(), "s", nil); err == nil || calls != 1 {
		t.Errorf("err=%v calls=%d, want an error after exactly 1 call", err, calls)
	}
}

func TestEvaluateGivesUpAfterMaxAttempts(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(429)
	}))
	defer server.Close()
	var sleeps []time.Duration
	client := testClient(server.URL, &sleeps)
	if _, err := client.Evaluate(context.Background(), "s", nil); err == nil || calls != client.MaxAttempts {
		t.Errorf("err=%v calls=%d, want an error after %d calls", err, calls, client.MaxAttempts)
	}
}
