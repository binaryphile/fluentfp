// Package sagas compile-checks the showcase entry for cockroachdb saga compensation.
package sagas

import (
	"context"

	"github.com/binaryphile/fluentfp/slice"
)

// --- stubs for the cockroachdb types referenced in the showcase ---

// ReplicationTarget stubs roachpb.ReplicationTarget.
type ReplicationTarget struct {
	StoreID int
}

// UUIDClient stubs the uuid package.
type UUIDClient struct{}

func (UUIDClient) MakeV4() string { return "" }

var uuid = UUIDClient{}

// Replica is the receiver type from the cockroachdb source.
type Replica struct{}

func (r Replica) addSnapshotLogTruncationConstraint(
	ctx context.Context, lockUUID string, b bool, storeID int,
) (struct{}, func()) {
	return struct{}{}, func() {}
}

// --- the fluentfp rewrite from docs/showcase.md (verbatim) ---

// call invokes a zero-argument function.
func call(fn func()) { fn() }

func (r *Replica) lockLearnerSnapshot(
	ctx context.Context, additions []ReplicationTarget,
) (unlock func()) {
	// acquireLock acquires a snapshot lock and returns its cleanup.
	acquireLock := func(addition ReplicationTarget) func() {
		lockUUID := uuid.MakeV4()
		_, cleanup := r.addSnapshotLogTruncationConstraint(
			ctx, lockUUID, true, addition.StoreID)
		return cleanup
	}

	var cleanups slice.Mapper[func()] = slice.Map(additions, acquireLock)
	return func() { cleanups.Each(call) }
}
