package agent

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ShutdownManager manages graceful shutdown of the application
type ShutdownManager struct {
	mu            sync.Mutex
	hooks         []ShutdownHook
	timeout       time.Duration
	shutdownChan  chan os.Signal
	ctx           context.Context
	cancel        context.CancelFunc
	shutdownOnce  sync.Once
	isShuttingDown bool
}

// ShutdownHook represents a function to be called during shutdown
type ShutdownHook func(ctx context.Context) error

// NewShutdownManager creates a new shutdown manager
func NewShutdownManager(timeout time.Duration) *ShutdownManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	sm := &ShutdownManager{
		hooks:        make([]ShutdownHook, 0),
		timeout:      timeout,
		shutdownChan: make(chan os.Signal, 1),
		ctx:          ctx,
		cancel:       cancel,
	}
	
	// Register signal handlers
	signal.Notify(sm.shutdownChan, 
		os.Interrupt,     // SIGINT (Ctrl+C)
		syscall.SIGTERM,  // SIGTERM (docker stop)
		syscall.SIGQUIT,  // SIGQUIT
	)
	
	return sm
}

// AddHook adds a shutdown hook
func (sm *ShutdownManager) AddHook(hook ShutdownHook) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.hooks = append(sm.hooks, hook)
}

// Context returns the shutdown context
func (sm *ShutdownManager) Context() context.Context {
	return sm.ctx
}

// IsShuttingDown returns true if shutdown is in progress
func (sm *ShutdownManager) IsShuttingDown() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.isShuttingDown
}

// WaitForShutdown waits for a shutdown signal and executes hooks
func (sm *ShutdownManager) WaitForShutdown() error {
	// Wait for shutdown signal
	sig := <-sm.shutdownChan
	
	if Logger != nil {
		Logger.InfoWithFields("shutdown-manager", "", 
			"Received shutdown signal", "", sig.String())
	}
	
	return sm.Shutdown()
}

// Shutdown initiates the graceful shutdown process
func (sm *ShutdownManager) Shutdown() error {
	var shutdownErr error
	
	sm.shutdownOnce.Do(func() {
		sm.mu.Lock()
		sm.isShuttingDown = true
		sm.mu.Unlock()
		
		if Logger != nil {
			Logger.Info("Starting graceful shutdown...")
		}
		
		// Create timeout context for shutdown hooks
		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), sm.timeout)
		defer timeoutCancel()
		
		// Execute shutdown hooks in reverse order
		for i := len(sm.hooks) - 1; i >= 0; i-- {
			hook := sm.hooks[i]
			
			if err := hook(timeoutCtx); err != nil {
				if Logger != nil {
					Logger.ErrorWithFields("shutdown-manager", "", 
						"Shutdown hook failed", err)
				}
				if shutdownErr == nil {
					shutdownErr = err
				}
			}
		}
		
		// Cancel the main context
		sm.cancel()
		
		if Logger != nil {
			if shutdownErr == nil {
				Logger.Info("Graceful shutdown completed successfully")
			} else {
				Logger.Error("Graceful shutdown completed with errors", shutdownErr)
			}
		}
	})
	
	return shutdownErr
}

// TestManager manages running tests with graceful shutdown support
type TestManager struct {
	mu           sync.RWMutex
	runningTests map[string]context.CancelFunc
	wg           sync.WaitGroup
	shutdown     *ShutdownManager
}

// NewTestManager creates a new test manager
func NewTestManager(shutdown *ShutdownManager) *TestManager {
	tm := &TestManager{
		runningTests: make(map[string]context.CancelFunc),
		shutdown:     shutdown,
	}
	
	// Register shutdown hook
	shutdown.AddHook(tm.shutdownHook)
	
	return tm
}

// StartTest starts a new test with cancellation support
func (tm *TestManager) StartTest(testID string, testFunc func(ctx context.Context) error) {
	tm.wg.Add(1)
	
	// Create cancellable context
	ctx, cancel := context.WithCancel(tm.shutdown.Context())
	
	// Register the test
	tm.mu.Lock()
	tm.runningTests[testID] = cancel
	tm.mu.Unlock()
	
	go func() {
		defer tm.wg.Done()
		defer func() {
			// Cleanup test registration
			tm.mu.Lock()
			delete(tm.runningTests, testID)
			tm.mu.Unlock()
		}()
		
		if err := testFunc(ctx); err != nil && err != context.Canceled {
			if Logger != nil {
				Logger.ErrorWithFields("test-manager", testID, 
					"Test execution failed", err)
			}
		}
	}()
}

// StopTest stops a specific test
func (tm *TestManager) StopTest(testID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	if cancel, exists := tm.runningTests[testID]; exists {
		cancel()
		delete(tm.runningTests, testID)
	}
}

// shutdownHook implements graceful shutdown for the test manager
func (tm *TestManager) shutdownHook(ctx context.Context) error {
	if Logger != nil {
		Logger.Info("Stopping all running tests...")
	}
	
	// Cancel all running tests
	tm.mu.Lock()
	runningCount := len(tm.runningTests)
	for testID, cancel := range tm.runningTests {
		if Logger != nil {
			Logger.DebugWithFields("test-manager", testID, "Cancelling test")
		}
		cancel()
	}
	tm.runningTests = make(map[string]context.CancelFunc)
	tm.mu.Unlock()
	
	if runningCount > 0 && Logger != nil {
		Logger.InfoWithFields("test-manager", "", 
			"Cancelled running tests", "", "")
	}
	
	// Wait for all tests to finish with timeout
	done := make(chan struct{})
	go func() {
		tm.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		if Logger != nil {
			Logger.Info("All tests stopped successfully")
		}
		return nil
	case <-ctx.Done():
		if Logger != nil {
			Logger.Warn("Timeout waiting for tests to stop")
		}
		return ctx.Err()
	}
}

// HTTPServerManager manages HTTP server with graceful shutdown
type HTTPServerManager struct {
	server *http.Server
}

// NewHTTPServerManager creates a new HTTP server manager
func NewHTTPServerManager(addr string, handler http.Handler) *HTTPServerManager {
	return &HTTPServerManager{
		server: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
}

// Start starts the HTTP server
func (hsm *HTTPServerManager) Start() error {
	if Logger != nil {
		Logger.InfoWithFields("http-server", hsm.server.Addr, 
			"Starting HTTP server", "", "")
	}
	
	if err := hsm.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully shuts down the HTTP server
func (hsm *HTTPServerManager) Shutdown(ctx context.Context) error {
	if Logger != nil {
		Logger.InfoWithFields("http-server", hsm.server.Addr, 
			"Shutting down HTTP server", "", "")
	}
	
	return hsm.server.Shutdown(ctx)
}

// ShutdownHook returns a shutdown hook for the HTTP server
func (hsm *HTTPServerManager) ShutdownHook() ShutdownHook {
	return hsm.Shutdown
}
