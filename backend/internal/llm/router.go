// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package llm

import (
	"context"
	"fmt"
)

// TaskType identifies the kind of LLM call, used by the router to select
// the appropriate model, provider, and default parameters.
type TaskType string

const (
	// TaskSummarize is used for generating summaries of fitness data.
	TaskSummarize TaskType = "summarize"
	// TaskPlan is used for structured planning tasks such as workout program generation.
	TaskPlan TaskType = "plan"
)

// Request is the caller's view of an LLM call. It carries content and semantic
// capabilities — not provider-specific details. Provider translation (e.g.
// ExtraBody injection for thinking mode) is the router's responsibility.
type Request struct {
	Messages []Message
	// Thinking enables extended reasoning if the provider supports it.
	// nil defers to the task or provider default.
	Thinking *bool
	// Temperature controls output randomness.
	// nil defers to the task or provider default.
	Temperature *float64
}

// Router selects a provider and dispatches LLM calls. Callers never interact
// with Client directly — all model selection, provider-specific parameter
// translation, and credential resolution happen inside the router.
type Router interface {
	Complete(ctx context.Context, task TaskType, req Request) (string, error)
}

// StaticRouter wraps a single Client and applies per-task defaults. It uses
// server-level configuration for all requests, ignoring any user or org
// credentials in ctx. This is the default implementation; a ContextualRouter
// can be substituted later to support per-user models and credentials.
type StaticRouter struct {
	client Client
}

// NewStaticRouter returns a Router backed by the given Client.
func NewStaticRouter(c Client) Router {
	return &StaticRouter{client: c}
}

// Complete dispatches the request to the configured client after applying
// task-level defaults for temperature and thinking mode.
func (r *StaticRouter) Complete(ctx context.Context, task TaskType, req Request) (string, error) {
	text, err := r.client.Complete(ctx, r.buildRequest(task, req))
	if err != nil {
		return "", fmt.Errorf("router(%s): %w", task, err)
	}
	return text, nil
}

// buildRequest translates a caller Request into the internal CompletionRequest,
// applying task-level defaults for any unset semantic fields and translating
// provider-specific parameters (e.g. thinking mode into ExtraBody).
func (r *StaticRouter) buildRequest(task TaskType, req Request) CompletionRequest {
	cr := CompletionRequest{
		Messages:    req.Messages,
		Temperature: req.Temperature,
	}
	if cr.Temperature == nil {
		cr.Temperature = defaultTemperature(task)
	}

	thinking := false // disabled by default; enable per-task as capabilities evolve
	if req.Thinking != nil {
		thinking = *req.Thinking
	}
	cr.ExtraBody = map[string]any{
		"enable_thinking": thinking,
	}
	return cr
}

// defaultTemperature returns the default temperature for a given task type.
func defaultTemperature(task TaskType) *float64 {
	switch task {
	case TaskSummarize:
		return Ptr(0.1)
	default:
		return nil
	}
}
