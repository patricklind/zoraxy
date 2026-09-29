package configstore

import (
	"context"
	"errors"
	"time"
)

// Follow applies complete revisions in order. applied is called only after a
// successful atomic activation; rejected revisions leave the old runtime live.
func Follow(ctx context.Context, store Store, activator Activator, after uint64, applied func(Revision), rejected func(Revision, error)) error {
	revisions, failures := store.Watch(ctx, after)
	for revisions != nil || failures != nil {
		select {
		case revision, ok := <-revisions:
			if !ok {
				revisions = nil
				continue
			}
			if err := activator.Activate(ctx, revision); err != nil {
				rejected(revision, err)
				continue
			}
			applied(revision)
		case err, ok := <-failures:
			if !ok {
				failures = nil
				continue
			}
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// FollowNode follows revisions and persists convergence state for one data
// node. An activation error is reported but the follower continues so a later
// valid revision can recover the node. A status-write error is returned because
// the control plane can no longer safely determine convergence.
func FollowNode(ctx context.Context, store Store, activator Activator, statuses NodeStatusStore, node NodeStatus) error {
	if node.NodeID == "" {
		return errors.New("node id is required")
	}
	if node.NodeRole == "" {
		return errors.New("node role is required")
	}
	if statuses == nil {
		return errors.New("node status store is required")
	}
	if err := statuses.UpsertNodeStatus(ctx, node); err != nil {
		return err
	}

	revisions, failures := store.Watch(ctx, node.AppliedRevision)
	for revisions != nil || failures != nil {
		select {
		case revision, ok := <-revisions:
			if !ok {
				revisions = nil
				continue
			}
			node.ConfigRevision = revision.ID
			node.LastError = ""
			node.UpdatedAt = time.Now().UTC()
			if err := statuses.UpsertNodeStatus(ctx, node); err != nil {
				return err
			}

			if err := activator.Activate(ctx, revision); err != nil {
				node.LastError = err.Error()
				node.UpdatedAt = time.Now().UTC()
				if reportErr := statuses.UpsertNodeStatus(ctx, node); reportErr != nil {
					return errors.Join(err, reportErr)
				}
				continue
			}

			node.AppliedRevision = revision.ID
			node.UpdatedAt = time.Now().UTC()
			if err := statuses.UpsertNodeStatus(ctx, node); err != nil {
				return err
			}
		case err, ok := <-failures:
			if !ok {
				failures = nil
				continue
			}
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
