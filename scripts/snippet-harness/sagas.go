//go:build ignore

// Package snippet is the verification harness for the sagas showcase
// entry in docs/showcase.md (cockroachdb saga-compensation pattern
// from replica_command.lockLearnerSnapshot). The snippet declares a
// package-level `call` helper and a method on *Replica; both go at
// package scope.
//
// The harness stubs the cockroachdb types referenced (ReplicationTarget,
// Replica with addSnapshotLogTruncationConstraint, uuid singleton with
// MakeV4) without pulling in the cockroachdb module.
//
// The `go:build ignore` constraint excludes this file from default
// `go build ./...`; scripts/check-snippets.py strips the constraint
// when assembling into the tmpdir.
package snippet

import (
	"context"

	"github.com/binaryphile/fluentfp/slice"
)

// ReplicationTarget stubs roachpb.ReplicationTarget — only StoreID
// is exercised by the snippet.
type ReplicationTarget struct {
	StoreID int
}

// UUIDClient stubs the uuid package's MakeV4 entry point.
type UUIDClient struct{}

func (UUIDClient) MakeV4() string { return "" }

var uuid = UUIDClient{}

// Replica is the receiver type from the cockroachdb source. The
// stubbed method returns the same `(struct{}, func())` shape the
// snippet's _, cleanup := destructure expects.
type Replica struct{}

func (r Replica) addSnapshotLogTruncationConstraint(
	ctx context.Context, lockUUID string, b bool, storeID int,
) (struct{}, func()) {
	return struct{}{}, func() {}
}

// __SNIPPET__
