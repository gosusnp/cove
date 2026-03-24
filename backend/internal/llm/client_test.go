// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(cfg Config, srv *httptest.Server) *openAICompatClient {
	cfg.BaseURL = srv.URL
	return &openAICompatClient{cfg: cfg, http: srv.Client()}
}

func TestComplete_success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIChatResponse{
			Choices: []openAIChoice{
				{Message: openAIMessage{Role: "assistant", Content: "Great session."}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient(Config{Model: "test-model"}, srv)
	got, err := c.Complete(context.Background(), CompletionRequest{
		Messages: []Message{{Role: "user", Content: "summarize"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Great session." {
		t.Errorf("got %q, want %q", got, "Great session.")
	}
}

func TestComplete_sendsAPIKey(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		resp := openAIChatResponse{Choices: []openAIChoice{{Message: openAIMessage{Content: "ok"}}}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient(Config{Model: "m", APIKey: "secret"}, srv)
	_, _ = c.Complete(context.Background(), CompletionRequest{})
	if gotAuth != "Bearer secret" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer secret")
	}
}

func TestComplete_omitsAuthHeaderWhenNoAPIKey(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		resp := openAIChatResponse{Choices: []openAIChoice{{Message: openAIMessage{Content: "ok"}}}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient(Config{Model: "m"}, srv)
	_, _ = c.Complete(context.Background(), CompletionRequest{})
	if gotAuth != "" {
		t.Errorf("expected no Authorization header, got %q", gotAuth)
	}
}

func TestComplete_modelOverride(t *testing.T) {
	var gotModel string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body openAIChatRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotModel = body.Model
		resp := openAIChatResponse{Choices: []openAIChoice{{Message: openAIMessage{Content: "ok"}}}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient(Config{Model: "default-model"}, srv)
	_, _ = c.Complete(context.Background(), CompletionRequest{Model: "override-model"})
	if gotModel != "override-model" {
		t.Errorf("model = %q, want %q", gotModel, "override-model")
	}
}

func TestComplete_nonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(Config{Model: "m"}, srv)
	_, err := c.Complete(context.Background(), CompletionRequest{})
	if err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}
}

func TestComplete_noChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIChatResponse{Choices: []openAIChoice{}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient(Config{Model: "m"}, srv)
	_, err := c.Complete(context.Background(), CompletionRequest{})
	if err == nil {
		t.Fatal("expected error for empty choices, got nil")
	}
}

func TestComplete_malformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{not valid`))
	}))
	defer srv.Close()

	c := newTestClient(Config{Model: "m"}, srv)
	_, err := c.Complete(context.Background(), CompletionRequest{})
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestBuildBody_noExtra(t *testing.T) {
	base := openAIChatRequest{Model: "gpt-4o", Messages: []openAIMessage{{Role: "user", Content: "hi"}}}
	b, err := buildBody(base, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["model"] != "gpt-4o" {
		t.Errorf("model = %v, want gpt-4o", got["model"])
	}
}

func TestBuildBody_extraOverridesField(t *testing.T) {
	base := openAIChatRequest{Model: "base-model"}
	b, err := buildBody(base, map[string]any{"model": "override-model", "enable_thinking": false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["model"] != "override-model" {
		t.Errorf("model = %v, want override-model", got["model"])
	}
	if got["enable_thinking"] != false {
		t.Errorf("enable_thinking = %v, want false", got["enable_thinking"])
	}
}
