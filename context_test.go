package generic_test

import (
	"context"
	"testing"
	"time"

	"github.com/agentflare-ai/go-generic"
)

func TestSubContextImplementsContext(t *testing.T) {
	// Test that SubContext implements context.Context interface
	var _ context.Context = (*generic.SubContext[context.Context])(nil)
}

func TestSubContextWithBackground(t *testing.T) {
	baseCtx := context.Background()
	subCtx := &generic.SubContext[context.Context]{
		Context: baseCtx,
	}

	// Test BaseContext returns the original context
	if got := subCtx.BaseContext(); got != baseCtx {
		t.Errorf("BaseContext() = %v, want %v", got, baseCtx)
	}

	// Test context methods are delegated
	deadline, ok := subCtx.Deadline()
	if ok {
		t.Error("Background context should not have deadline")
	}
	if deadline.IsZero() == false {
		t.Error("Background context deadline should be zero")
	}

	select {
	case <-subCtx.Done():
		t.Error("Background context should not be done")
	default:
		// Expected
	}

	if err := subCtx.Err(); err != nil {
		t.Errorf("Background context should not have error, got %v", err)
	}

	if val := subCtx.Value("key"); val != nil {
		t.Errorf("Background context should return nil for any key, got %v", val)
	}
}

func TestSubContextWithValue(t *testing.T) {
	type testKey string
	baseCtx := context.WithValue(context.Background(), testKey("key"), "value")
	subCtx := &generic.SubContext[context.Context]{
		Context: baseCtx,
	}

	// Test BaseContext returns the original context
	if got := subCtx.BaseContext(); got != baseCtx {
		t.Errorf("BaseContext() = %v, want %v", got, baseCtx)
	}

	// Test Value method works
	if val := subCtx.Value(testKey("key")); val != "value" {
		t.Errorf("Value(%q) = %v, want %q", "key", val, "value")
	}

	if val := subCtx.Value(testKey("nonexistent")); val != nil {
		t.Errorf("Value(%q) = %v, want nil", "nonexistent", val)
	}
}

func TestSubContextWithCancel(t *testing.T) {
	baseCtx, cancel := context.WithCancel(context.Background())
	subCtx := &generic.SubContext[context.Context]{
		Context: baseCtx,
	}

	// Test BaseContext returns the original context
	if got := subCtx.BaseContext(); got != baseCtx {
		t.Errorf("BaseContext() = %v, want %v", got, baseCtx)
	}

	// Test Done channel works
	select {
	case <-subCtx.Done():
		t.Error("Context should not be done before cancel")
	default:
		// Expected
	}

	if err := subCtx.Err(); err != nil {
		t.Errorf("Context should not have error before cancel, got %v", err)
	}

	// Cancel and test
	cancel()

	select {
	case <-time.After(10 * time.Millisecond):
		t.Error("Context should be done after cancel")
	case <-subCtx.Done():
		// Expected
	}

	if err := subCtx.Err(); err != context.Canceled {
		t.Errorf("Context should have Canceled error after cancel, got %v", err)
	}
}

func TestSubContextWithDeadline(t *testing.T) {
	deadline := time.Now().Add(50 * time.Millisecond)
	baseCtx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	subCtx := &generic.SubContext[context.Context]{
		Context: baseCtx,
	}

	// Test BaseContext returns the original context
	if got := subCtx.BaseContext(); got != baseCtx {
		t.Errorf("BaseContext() = %v, want %v", got, baseCtx)
	}

	// Test Deadline method
	dl, ok := subCtx.Deadline()
	if !ok {
		t.Error("Context with deadline should return ok=true")
	}
	if !dl.Equal(deadline) {
		t.Errorf("Deadline() = %v, want %v", dl, deadline)
	}

	// Test that it gets canceled when deadline passes
	time.Sleep(60 * time.Millisecond)

	select {
	case <-subCtx.Done():
		// Expected
	default:
		t.Error("Context should be done after deadline")
	}

	if err := subCtx.Err(); err != context.DeadlineExceeded {
		t.Errorf("Context should have DeadlineExceeded error, got %v", err)
	}
}

func TestSubContextWithTimeout(t *testing.T) {
	baseCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	subCtx := &generic.SubContext[context.Context]{
		Context: baseCtx,
	}

	// Test BaseContext returns the original context
	if got := subCtx.BaseContext(); got != baseCtx {
		t.Errorf("BaseContext() = %v, want %v", got, baseCtx)
	}

	// Test Deadline method
	dl, ok := subCtx.Deadline()
	if !ok {
		t.Error("Context with timeout should have deadline")
	}
	if dl.IsZero() {
		t.Error("Deadline should not be zero")
	}

	// Test that it gets canceled when timeout passes
	time.Sleep(60 * time.Millisecond)

	select {
	case <-subCtx.Done():
		// Expected
	default:
		t.Error("Context should be done after timeout")
	}

	if err := subCtx.Err(); err != context.DeadlineExceeded {
		t.Errorf("Context should have DeadlineExceeded error, got %v", err)
	}
}

func TestSubContextGenericBehavior(t *testing.T) {
	// Test with context.Background
	baseBg := context.Background()
	subBg := &generic.SubContext[context.Context]{
		Context: baseBg,
	}

	if subBg.BaseContext() != baseBg {
		t.Error("Generic context.Background test failed")
	}

	// Test with context.TODO
	baseTodo := context.TODO()
	subTodo := &generic.SubContext[context.Context]{
		Context: baseTodo,
	}

	if subTodo.BaseContext() != baseTodo {
		t.Error("Generic context.TODO test failed")
	}

	// Test with WithValue
	type customKey string
	baseWithValue := context.WithValue(context.Background(), customKey("custom"), "custom-value")
	subWithValue := &generic.SubContext[context.Context]{
		Context: baseWithValue,
	}

	if subWithValue.BaseContext() != baseWithValue {
		t.Error("Generic context.WithValue test failed")
	}

	if val := subWithValue.Value(customKey("custom")); val != "custom-value" {
		t.Errorf("Value retrieval failed: got %v, want %q", val, "custom-value")
	}
}
