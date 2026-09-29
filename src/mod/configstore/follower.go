package configstore

import "context"

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
