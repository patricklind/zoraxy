package configstore

import (
	"context"
	"sync"
	"testing"
	"time"
)

type scriptedLeaseStore struct {
	mu       sync.Mutex
	results  []bool
	acquires chan struct{}
}

func (s *scriptedLeaseStore) AcquireCertificateControllerLease(context.Context, string, time.Duration) (ControllerLease, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	acquired := false
	if len(s.results) > 0 {
		acquired = s.results[0]
		s.results = s.results[1:]
	}
	select {
	case s.acquires <- struct{}{}:
	default:
	}
	return ControllerLease{}, acquired, nil
}

func TestCertificateControllerCancelsWorkerWhenLeaseIsLost(t *testing.T) {
	store := &scriptedLeaseStore{results: []bool{true, false}, acquires: make(chan struct{}, 4)}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{})
	stopped := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- RunCertificateControllerLease(ctx, store, "holder", 6*time.Second, func(workerContext context.Context) error {
			close(started)
			<-workerContext.Done()
			close(stopped)
			return workerContext.Err()
		})
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("worker did not start after lease acquisition")
	}
	select {
	case <-stopped:
	case <-time.After(3 * time.Second):
		t.Fatal("worker was not cancelled after lease loss")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("lease runner did not stop")
	}
}
