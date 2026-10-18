// client.go
//
// Responsibility: send one state and its questions to the JEV API and return the raw response.
// It owns the network call and the retry policy, and knows nothing about what the questions mean.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	jevEndpoint        = "https://api.typesafe.ai/v1/systemone"
	jevModel           = "jev-latest"
	defaultMaxAttempts = 5
	defaultBaseDelay   = time.Second
)

// JEVClient is configured at the edge; Sleep is injectable so retries are testable.
type JEVClient struct {
	APIKey      string
	Endpoint    string
	HTTP        *http.Client
	MaxAttempts int
	BaseDelay   time.Duration
	Sleep       func(time.Duration)
}

func NewJEVClient(apiKey string) JEVClient {
	return JEVClient{
		APIKey:      apiKey,
		Endpoint:    jevEndpoint,
		HTTP:        http.DefaultClient,
		MaxAttempts: defaultMaxAttempts,
		BaseDelay:   defaultBaseDelay,
		Sleep:       time.Sleep,
	}
}

type jevRequest struct {
	State     any            `json:"state"`
	Model     string         `json:"model"`
	Questions map[string]any `json:"questions"`
}

// jevResponse keeps answers raw; results.go validates and types them.
type jevResponse struct {
	Model   string                     `json:"model"`
	Answers map[string]json.RawMessage `json:"answers"`
	Usage   struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Evaluate asks every question about one state. Rate limits (429) and overload (529) are
// retried with exponential backoff; any other failure returns immediately.
func (c JEVClient) Evaluate(ctx context.Context, state any, questions map[string]any) (jevResponse, error) {
	body, err := json.Marshal(jevRequest{State: state, Model: jevModel, Questions: questions})
	if err != nil {
		return jevResponse{}, err
	}
	for attempt := 0; ; attempt++ {
		response, retryable, err := c.post(ctx, body)
		if err == nil {
			return response, nil
		}
		if !retryable || attempt+1 >= c.MaxAttempts {
			return jevResponse{}, err
		}
		c.Sleep(backoffDelay(c.BaseDelay, attempt))
	}
}

func (c JEVClient) post(ctx context.Context, body []byte) (response jevResponse, retryable bool, err error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(body))
	if err != nil {
		return jevResponse{}, false, err
	}
	request.Header.Set("Authorization", "Bearer "+c.APIKey)
	request.Header.Set("Content-Type", "application/json")
	httpResponse, err := c.HTTP.Do(request)
	if err != nil {
		return jevResponse{}, false, err
	}
	defer httpResponse.Body.Close()
	payload, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return jevResponse{}, false, err
	}
	if httpResponse.StatusCode != http.StatusOK {
		statusErr := fmt.Errorf("jev: status %d: %s", httpResponse.StatusCode, payload)
		return jevResponse{}, isRetryableStatus(httpResponse.StatusCode), statusErr
	}
	return response, false, json.Unmarshal(payload, &response)
}

func isRetryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || status == 529
}

// backoffDelay doubles with each attempt: base, 2*base, 4*base, ...
func backoffDelay(base time.Duration, attempt int) time.Duration {
	return base << attempt
}
