package configstore

import (
	"context"
	"errors"
	"time"
)

type CertificateLeaseStore interface {
	AcquireCertificateControllerLease(context.Context, string, time.Duration) (ControllerLease, bool, error)
}

// RunCertificateControllerLease maintains leadership and gives the worker a
// context that is cancelled immediately when renewal fails or ownership is
// lost. The worker must pass that context to every external operation.
func RunCertificateControllerLease(ctx context.Context, store CertificateLeaseStore, holderID string, ttl time.Duration, worker func(context.Context) error) error {
	if store == nil || worker == nil {
		return errors.New("certificate lease store and worker are required")
	}
	if ttl < 5*time.Second || ttl > 5*time.Minute {
		return errors.New("controller lease ttl must be between 5 seconds and 5 minutes")
	}
	interval := ttl / 3
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var cancelLeadership context.CancelFunc
	workerRunning := false
	workerDone := make(chan error, 1)
	defer func() {
		if cancelLeadership != nil {
			cancelLeadership()
		}
	}()

	acquire := func() error {
		leaseContext, cancel := context.WithTimeout(ctx, interval)
		defer cancel()
		_, acquired, err := store.AcquireCertificateControllerLease(leaseContext, holderID, ttl)
		if err != nil || !acquired {
			if cancelLeadership != nil {
				cancelLeadership()
				cancelLeadership = nil
			}
			return err
		}
		if !workerRunning {
			leadershipContext, cancel := context.WithCancel(ctx)
			cancelLeadership = cancel
			workerRunning = true
			go func() { workerDone <- worker(leadershipContext) }()
		}
		return nil
	}

	if err := acquire(); err != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-workerDone:
			workerRunning = false
			if cancelLeadership != nil {
				cancelLeadership()
				cancelLeadership = nil
			}
			if err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
		case <-ticker.C:
			_ = acquire()
		}
	}
}
