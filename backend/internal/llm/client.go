// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

// Package llm provides a generic interface and an OpenAI-compatible implementation
// for calling large language models. It is provider-agnostic: any endpoint that
// speaks the OpenAI chat completions API (OpenAI, KubeAI, Anthropic compat, etc.)
// works with the same client.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client is the provider-agnostic interface for calling an LLM.
type Client interface {
	// Complete sends the request and returns the full response as a string.
	Complete(ctx context.Context, req CompletionRequest) (string, error)
}

// Ptr returns a pointer to v. Useful for passing literal values to pointer fields.
func Ptr[T any](v T) *T {
	return &v
}

// Message is a single turn in a conversation.
type Message struct {
	// Role is one of "system", "user", or "assistant".
	Role    string
	Content string
}

// CompletionRequest is a provider-agnostic chat completion request.
type CompletionRequest struct {
	Messages []Message
	// Model overrides the client's configured default when non-empty.
	Model       string
	Temperature *float64
	// ExtraBody is merged into the JSON request body last, allowing
	// provider-specific parameters (e.g. Anthropic thinking mode, top_p,
	// stop sequences) without breaking the generic interface.
	// Keys in ExtraBody override any field already set by the client.
	ExtraBody map[string]any
}

// Config holds the parameters for an OpenAI-compatible LLM client.
type Config struct {
	// BaseURL is the root of the chat completions API, e.g.
	// "https://api.openai.com/v1" or an internal KubeAI endpoint.
	BaseURL string
	// APIKey is sent as a Bearer token. May be empty for unauthenticated endpoints.
	APIKey string
	// Model is the default model name used when CompletionRequest.Model is empty.
	Model string
	// Timeout overrides the default HTTP client timeout. Useful when the
	// backend may cold-start (e.g. KubeAI spinning up a model pod).
	// Defaults to 5 minutes when zero.
	Timeout time.Duration
	// Debug writes the raw response body to stderr before parsing.
	// Useful for inspecting provider-specific fields such as thinking blocks.
	Debug bool
}

// openAICompatClient calls any OpenAI-compatible chat completions endpoint.
type openAICompatClient struct {
	cfg  Config
	http *http.Client
}

// NewOpenAICompatClient returns a Client that speaks the OpenAI chat completions
// API. It works with any compatible provider: OpenAI, KubeAI, Anthropic's
// compatibility endpoint, and others.
func NewOpenAICompatClient(cfg Config) Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	return &openAICompatClient{
		cfg:  cfg,
		http: &http.Client{Timeout: timeout},
	}
}

// openAIChatRequest is the body sent to POST /chat/completions.
type openAIChatRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature *float64        `json:"temperature,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIChatResponse is the subset of the chat completions response we read.
type openAIChatResponse struct {
	Choices []openAIChoice `json:"choices"`
}

type openAIChoice struct {
	Message openAIMessage `json:"message"`
}

// buildBody marshals base into JSON, then merges extra over the top.
// Keys in extra take precedence, allowing provider-specific fields to be
// injected without altering the typed struct.
func buildBody(base openAIChatRequest, extra map[string]any) ([]byte, error) {
	if len(extra) == 0 {
		return json.Marshal(base)
	}
	b, err := json.Marshal(base)
	if err != nil {
		return nil, err
	}
	var merged map[string]any
	if err := json.Unmarshal(b, &merged); err != nil {
		return nil, err
	}
	for k, v := range extra {
		merged[k] = v
	}
	return json.Marshal(merged)
}

// Complete sends the request to the configured chat completions endpoint and
// returns the assistant message content.
func (c *openAICompatClient) Complete(ctx context.Context, req CompletionRequest) (string, error) {
	model := req.Model
	if model == "" {
		model = c.cfg.Model
	}

	msgs := make([]openAIMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = openAIMessage(m)
	}

	body, err := buildBody(openAIChatRequest{
		Model:       model,
		Messages:    msgs,
		Temperature: req.Temperature,
	}, req.ExtraBody)
	if err != nil {
		return "", fmt.Errorf("llm complete request: %w", err)
	}

	endpoint := c.cfg.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("llm complete request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("llm complete: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("llm complete: status %d from %s: %s", resp.StatusCode, endpoint, errBody)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("llm complete read: %w", err)
	}
	if c.cfg.Debug {
		fmt.Fprintf(os.Stderr, "llm response: %s\n", respBody)
	}

	var result openAIChatResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("llm complete decode: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("llm complete: no choices in response")
	}

	return result.Choices[0].Message.Content, nil
}
