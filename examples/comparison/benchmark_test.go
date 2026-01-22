package main

import (
	"testing"

	"github.com/ahmetb/go-linq/v3"
	"github.com/binaryphile/fluentfp/slice"
	u "github.com/rjNemo/underscore"
	"github.com/samber/lo"
	"github.com/thoas/go-funk"
)

// User definition for benchmarks (main.go has //go:build ignore).
type User struct {
	active bool
	name   string
}

func (u User) Name() string   { return u.name }
func (u User) IsActive() bool { return u.active }

// Benchmarks compare filter+map operations across FP libraries.
// All tests use the same data: 1000 users, ~50% active.

var benchUsers []User

func init() {
	benchUsers = make([]User, 1000)
	for i := range benchUsers {
		benchUsers[i] = User{
			name:   "User" + string(rune('A'+i%26)),
			active: i%2 == 0, // 50% active
		}
	}
}

// BenchmarkFluentfp benchmarks fluentfp filter+map.
func BenchmarkFluentfp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = slice.From(benchUsers).
			KeepIf(User.IsActive).
			ToString(User.Name)
	}
}

// BenchmarkLo benchmarks samber/lo filter+map.
func BenchmarkLo(b *testing.B) {
	userIsActive := func(u User, _ int) bool { return u.IsActive() }
	getName := func(u User, _ int) string { return u.Name() }

	for i := 0; i < b.N; i++ {
		actives := lo.Filter(benchUsers, userIsActive)
		_ = lo.Map(actives, getName)
	}
}

// BenchmarkGoFunk benchmarks thoas/go-funk filter+map.
func BenchmarkGoFunk(b *testing.B) {
	for i := 0; i < b.N; i++ {
		actives := funk.Filter(benchUsers, User.IsActive).([]User)
		_ = funk.Map(actives, User.Name).([]string)
	}
}

// BenchmarkGoLinq benchmarks ahmetb/go-linq filter+map.
func BenchmarkGoLinq(b *testing.B) {
	userIsActive := func(user any) bool { return user.(User).IsActive() }
	name := func(user any) any { return user.(User).Name() }

	for i := 0; i < b.N; i++ {
		var names []any
		linq.From(benchUsers).
			Where(userIsActive).
			Select(name).
			ToSlice(&names)
	}
}

// BenchmarkUnderscore benchmarks rjNemo/underscore filter+map.
func BenchmarkUnderscore(b *testing.B) {
	for i := 0; i < b.N; i++ {
		actives := u.Filter(benchUsers, User.IsActive)
		_ = u.Map(actives, User.Name)
	}
}

// BenchmarkLoop benchmarks pre-allocated loop for baseline.
func BenchmarkLoop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		actives := make([]User, 0, len(benchUsers))
		for _, user := range benchUsers {
			if user.IsActive() {
				actives = append(actives, user)
			}
		}
		names := make([]string, len(actives))
		for j, user := range actives {
			names[j] = user.Name()
		}
		_ = names
	}
}

// === FILTER ONLY ===

// BenchmarkFluentfp_FilterOnly benchmarks fluentfp filter.
func BenchmarkFluentfp_FilterOnly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = slice.From(benchUsers).KeepIf(User.IsActive)
	}
}

// BenchmarkLo_FilterOnly benchmarks samber/lo filter.
func BenchmarkLo_FilterOnly(b *testing.B) {
	userIsActive := func(u User, _ int) bool { return u.IsActive() }
	for i := 0; i < b.N; i++ {
		_ = lo.Filter(benchUsers, userIsActive)
	}
}

// BenchmarkGoFunk_FilterOnly benchmarks thoas/go-funk filter.
func BenchmarkGoFunk_FilterOnly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = funk.Filter(benchUsers, User.IsActive).([]User)
	}
}

// BenchmarkUnderscore_FilterOnly benchmarks rjNemo/underscore filter.
func BenchmarkUnderscore_FilterOnly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = u.Filter(benchUsers, User.IsActive)
	}
}

// BenchmarkLoop_FilterOnly benchmarks pre-allocated filter loop.
func BenchmarkLoop_FilterOnly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		actives := make([]User, 0, len(benchUsers))
		for _, user := range benchUsers {
			if user.IsActive() {
				actives = append(actives, user)
			}
		}
		_ = actives
	}
}

// === MAP ONLY ===

// BenchmarkFluentfp_MapOnly benchmarks fluentfp map.
func BenchmarkFluentfp_MapOnly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = slice.From(benchUsers).ToString(User.Name)
	}
}

// BenchmarkLo_MapOnly benchmarks samber/lo map.
func BenchmarkLo_MapOnly(b *testing.B) {
	getName := func(u User, _ int) string { return u.Name() }
	for i := 0; i < b.N; i++ {
		_ = lo.Map(benchUsers, getName)
	}
}

// BenchmarkGoFunk_MapOnly benchmarks thoas/go-funk map.
func BenchmarkGoFunk_MapOnly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = funk.Map(benchUsers, User.Name).([]string)
	}
}

// BenchmarkUnderscore_MapOnly benchmarks rjNemo/underscore map.
func BenchmarkUnderscore_MapOnly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = u.Map(benchUsers, User.Name)
	}
}

// BenchmarkLoop_MapOnly benchmarks hand-written map loop.
func BenchmarkLoop_MapOnly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		names := make([]string, len(benchUsers))
		for j, user := range benchUsers {
			names[j] = user.Name()
		}
		_ = names
	}
}
