package configstore

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Candidate is a complete, validated runtime configuration built off-path.
// Close releases resources owned by a candidate that was rejected or retired.
type Candidate interface {
	Close() error
}

// CandidateBuilder parses, validates and prepares a complete revision without
// changing the currently serving runtime.
type CandidateBuilder interface {
	Build(ctx context.Context, revision Revision) (Candidate, error)
}

// CandidateRuntime atomically replaces the serving configuration. If Swap
// returns an error, the old configuration must still be serving and previous
// must be nil.
type CandidateRuntime interface {
	Swap(ctx context.Context, candidate Candidate) (previous Candidate, err error)
}

// RetireErrorHandler receives cleanup failures after a successful swap. Such
// failures are operationally important, but cannot turn an already completed
// activation into a rejection.
type RetireErrorHandler func(error)

// AtomicActivator implements the configstore Activator contract while keeping
// construction and validation outside the serving runtime.
type AtomicActivator struct {
	builder       CandidateBuilder
	runtime       CandidateRuntime
	onRetireError RetireErrorHandler
	mu            sync.Mutex
}

func NewAtomicActivator(builder CandidateBuilder, runtime CandidateRuntime, onRetireError RetireErrorHandler) (*AtomicActivator, error) {
	if builder == nil {
		return nil, errors.New("candidate builder is required")
	}
	if runtime == nil {
		return nil, errors.New("candidate runtime is required")
	}
	return &AtomicActivator{builder: builder, runtime: runtime, onRetireError: onRetireError}, nil
}

func (a *AtomicActivator) Activate(ctx context.Context, revision Revision) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	candidate, err := a.builder.Build(ctx, revision)
	if err != nil {
		return fmt.Errorf("build revision %d: %w", revision.ID, err)
	}
	if candidate == nil {
		return fmt.Errorf("build revision %d: builder returned a nil candidate", revision.ID)
	}

	previous, err := a.runtime.Swap(ctx, candidate)
	if err != nil {
		closeErr := candidate.Close()
		if closeErr != nil {
			return errors.Join(fmt.Errorf("activate revision %d: %w", revision.ID, err),
				fmt.Errorf("discard rejected candidate: %w", closeErr))
		}
		return fmt.Errorf("activate revision %d: %w", revision.ID, err)
	}

	if previous != nil {
		if err := previous.Close(); err != nil && a.onRetireError != nil {
			a.onRetireError(fmt.Errorf("retire previous runtime after revision %d: %w", revision.ID, err))
		}
	}
	return nil
}
