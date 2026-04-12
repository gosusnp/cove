// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package workers

import (
	"context"
	"testing"
	"time"
)

type stubRefresher struct {
	called chan struct{}
	err    error
}

func newStubRefresher() *stubRefresher {
	return &stubRefresher{called: make(chan struct{}, 64)}
}

func (s *stubRefresher) RefreshTimeout(_ string) error {
	s.called <- struct{}{}
	return s.err
}

func waitForN(t *testing.T, ch <-chan struct{}, n int, deadline time.Duration) {
	t.Helper()
	timeout := time.After(deadline)
	for range n {
		select {
		case <-ch:
		case <-timeout:
			t.Fatalf("timed out waiting for RefreshTimeout call %d/%d", n, n)
		}
	}
}

func TestStartHeartbeat_CallsRefreshOnInterval(t *testing.T) {
	r := newStubRefresher()
	stop, _ := startHeartbeat(context.Background(), r, 20*time.Millisecond, 30*time.Millisecond, "test")
	defer stop()

	waitForN(t, r.called, 2, 500*time.Millisecond)
}

func TestStartHeartbeat_StopsOnStop(t *testing.T) {
	r := newStubRefresher()
	stop, done := startHeartbeat(context.Background(), r, 20*time.Millisecond, 30*time.Millisecond, "test")
	stop()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("heartbeat goroutine did not stop after stop()")
	}
}

func TestStartHeartbeat_StopsOnContextCancel(t *testing.T) {
	r := newStubRefresher()
	ctx, cancel := context.WithCancel(context.Background())
	_, done := startHeartbeat(ctx, r, 20*time.Millisecond, 30*time.Millisecond, "test")
	cancel()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("heartbeat goroutine did not stop after context cancel")
	}
}
