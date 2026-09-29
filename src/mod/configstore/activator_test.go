package configstore

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type testCandidate struct {
	name     string
	closeErr error
	events   *[]string
}

func (c *testCandidate) Close() error {
	*c.events = append(*c.events, "close:"+c.name)
	return c.closeErr
}

type testCandidateBuilder struct {
	candidate Candidate
	err       error
	events    *[]string
}

func (b testCandidateBuilder) Build(_ context.Context, _ Revision) (Candidate, error) {
	*b.events = append(*b.events, "build")
	return b.candidate, b.err
}

type testCandidateRuntime struct {
	previous Candidate
	err      error
	events   *[]string
}

func (r testCandidateRuntime) Swap(_ context.Context, _ Candidate) (Candidate, error) {
	*r.events = append(*r.events, "swap")
	return r.previous, r.err
}

func TestAtomicActivatorBuildsSwapsThenRetires(t *testing.T) {
	events := []string{}
	old := &testCandidate{name: "old", events: &events}
	next := &testCandidate{name: "next", events: &events}
	activator, err := NewAtomicActivator(
		testCandidateBuilder{candidate: next, events: &events},
		testCandidateRuntime{previous: old, events: &events},
		func(err error) { t.Fatalf("unexpected retire error: %v", err) },
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := activator.Activate(context.Background(), Revision{ID: 7}); err != nil {
		t.Fatal(err)
	}
	if want := []string{"build", "swap", "close:old"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestAtomicActivatorDiscardsCandidateWhenSwapFails(t *testing.T) {
	events := []string{}
	swapErr := errors.New("runtime rejected candidate")
	next := &testCandidate{name: "next", events: &events}
	activator, err := NewAtomicActivator(
		testCandidateBuilder{candidate: next, events: &events},
		testCandidateRuntime{err: swapErr, events: &events}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	err = activator.Activate(context.Background(), Revision{ID: 8})
	if !errors.Is(err, swapErr) {
		t.Fatalf("error = %v, want %v", err, swapErr)
	}
	if want := []string{"build", "swap", "close:next"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestAtomicActivatorDoesNotSwapInvalidCandidate(t *testing.T) {
	events := []string{}
	buildErr := errors.New("listener conflict")
	activator, err := NewAtomicActivator(
		testCandidateBuilder{err: buildErr, events: &events},
		testCandidateRuntime{events: &events}, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	err = activator.Activate(context.Background(), Revision{ID: 9})
	if !errors.Is(err, buildErr) {
		t.Fatalf("error = %v, want %v", err, buildErr)
	}
	if want := []string{"build"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestAtomicActivatorReportsRetireFailureWithoutRejectingAppliedRevision(t *testing.T) {
	events := []string{}
	retireErr := errors.New("old runtime cleanup failed")
	old := &testCandidate{name: "old", closeErr: retireErr, events: &events}
	next := &testCandidate{name: "next", events: &events}
	var reported error
	activator, err := NewAtomicActivator(
		testCandidateBuilder{candidate: next, events: &events},
		testCandidateRuntime{previous: old, events: &events},
		func(err error) { reported = err },
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := activator.Activate(context.Background(), Revision{ID: 10}); err != nil {
		t.Fatalf("successful swap was reported as rejected: %v", err)
	}
	if !errors.Is(reported, retireErr) {
		t.Fatalf("reported error = %v, want %v", reported, retireErr)
	}
}
