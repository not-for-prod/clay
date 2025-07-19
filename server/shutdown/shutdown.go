package shutdown

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"time"
)

// Manager handles graceful application shutdown.
type Manager struct {
	mu        sync.RWMutex
	completed sync.Once
	finished  chan struct{}
	handlers  []Handler
	timeout   time.Duration
}

// Handler represents a cleanup function.
type Handler func(ctx context.Context) error

// defaultManager is the package-level shutdown manager.
var defaultManager = NewManager()

// Register adds cleanup handlers to the default manager.
func Register(handlers ...Handler) {
	defaultManager.Register(handlers...)
}

// Shutdown triggers graceful shutdown of all registered handlers.
func Shutdown(ctx context.Context) error {
	return defaultManager.Shutdown(ctx)
}

// AwaitTermination blocks until shutdown is complete.
func AwaitTermination() {
	defaultManager.AwaitTermination()
}

// NewManager creates a new shutdown manager.
func NewManager(signals ...os.Signal) *Manager {
	m := &Manager{
		finished: make(chan struct{}),
		timeout:  30 * time.Second,
	}

	if len(signals) > 0 {
		go m.handleSignals(signals...)
	}

	return m
}

// Register adds cleanup handlers to the manager.
func (m *Manager) Register(handlers ...Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, handlers...)
}

// SetTimeout configures the shutdown timeout.
func (m *Manager) SetTimeout(timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.timeout = timeout
}

// Shutdown executes all registered handlers with timeout.
func (m *Manager) Shutdown(ctx context.Context) error {
	var shutdownErr error

	m.completed.Do(func() {
		defer close(m.finished)

		m.mu.RLock()
		currentHandlers := make([]Handler, len(m.handlers))
		copy(currentHandlers, m.handlers)
		timeout := m.timeout
		m.mu.RUnlock()

		if len(currentHandlers) == 0 {
			return
		}

		// Only if ctx doesn't have a timeout, we create our own
		var cancel context.CancelFunc
		if _, hasDeadline := ctx.Deadline(); !hasDeadline {
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		results := make(chan error, len(currentHandlers))

		// Execute handlers concurrently
		for _, handler := range currentHandlers {
			go func(h Handler) {
				results <- h(ctx)
			}(handler)
		}

		// Collect results
		var errs []error
		for i := 0; i < len(currentHandlers); i++ {
			if err := <-results; err != nil {
				errs = append(errs, err)
			}
		}

		if len(errs) > 0 {
			shutdownErr = errors.Join(errs...)
		}
	})

	return shutdownErr
}

// AwaitTermination blocks until shutdown completes.
func (m *Manager) AwaitTermination() {
	<-m.finished
}

// handleSignals listens for OS signals and triggers shutdown.
func (m *Manager) handleSignals(signals ...os.Signal) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, signals...)

	<-sigChan
	signal.Stop(sigChan)

	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	m.Shutdown(ctx)
}
