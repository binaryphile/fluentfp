package slice

import (
	"strconv"
	"testing"
)

// Test data setup
type benchUser struct {
	ID     int
	Name   string
	Active bool
}

func (u benchUser) IsActive() bool  { return u.Active }
func (u benchUser) GetName() string { return u.Name }

func makeUsers(n int) []benchUser {
	users := make([]benchUser, n)
	for i := 0; i < n; i++ {
		users[i] = benchUser{
			ID:     i,
			Name:   "user" + strconv.Itoa(i),
			Active: i%2 == 0, // 50% active
		}
	}
	return users
}

// Case 1: Filter only - KeepIf vs manual loop

func BenchmarkFilter_Loop_100(b *testing.B)   { benchFilterLoop(b, 100) }
func BenchmarkFilter_Loop_1000(b *testing.B)  { benchFilterLoop(b, 1000) }
func BenchmarkFilter_Loop_10000(b *testing.B) { benchFilterLoop(b, 10000) }

func BenchmarkFilter_Chain_100(b *testing.B)   { benchFilterChain(b, 100) }
func BenchmarkFilter_Chain_1000(b *testing.B)  { benchFilterChain(b, 1000) }
func BenchmarkFilter_Chain_10000(b *testing.B) { benchFilterChain(b, 10000) }

func benchFilterLoop(b *testing.B, n int) {
	users := makeUsers(n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result []benchUser
		for _, u := range users {
			if u.Active {
				result = append(result, u)
			}
		}
		_ = result
	}
}

func benchFilterChain(b *testing.B, n int) {
	users := makeUsers(n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := From(users).KeepIf(benchUser.IsActive)
		_ = result
	}
}

// Case 2: Filter + Map - KeepIf + ToString vs fused loop

func BenchmarkFilterMap_Loop_100(b *testing.B)   { benchFilterMapLoop(b, 100) }
func BenchmarkFilterMap_Loop_1000(b *testing.B)  { benchFilterMapLoop(b, 1000) }
func BenchmarkFilterMap_Loop_10000(b *testing.B) { benchFilterMapLoop(b, 10000) }

func BenchmarkFilterMap_Chain_100(b *testing.B)   { benchFilterMapChain(b, 100) }
func BenchmarkFilterMap_Chain_1000(b *testing.B)  { benchFilterMapChain(b, 1000) }
func BenchmarkFilterMap_Chain_10000(b *testing.B) { benchFilterMapChain(b, 10000) }

func benchFilterMapLoop(b *testing.B, n int) {
	users := makeUsers(n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result []string
		for _, u := range users {
			if u.Active {
				result = append(result, u.Name)
			}
		}
		_ = result
	}
}

func benchFilterMapChain(b *testing.B, n int) {
	users := makeUsers(n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := From(users).KeepIf(benchUser.IsActive).ToString(benchUser.GetName)
		_ = result
	}
}

// Case 3: Filter + Map + Count - chain vs fused loop

func BenchmarkFilterMapCount_Loop_100(b *testing.B)   { benchFilterMapCountLoop(b, 100) }
func BenchmarkFilterMapCount_Loop_1000(b *testing.B)  { benchFilterMapCountLoop(b, 1000) }
func BenchmarkFilterMapCount_Loop_10000(b *testing.B) { benchFilterMapCountLoop(b, 10000) }

func BenchmarkFilterMapCount_Chain_100(b *testing.B)   { benchFilterMapCountChain(b, 100) }
func BenchmarkFilterMapCount_Chain_1000(b *testing.B)  { benchFilterMapCountChain(b, 1000) }
func BenchmarkFilterMapCount_Chain_10000(b *testing.B) { benchFilterMapCountChain(b, 10000) }

func benchFilterMapCountLoop(b *testing.B, n int) {
	users := makeUsers(n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		for _, u := range users {
			if u.Active {
				_ = u.Name // simulate map work
				count++
			}
		}
		_ = count
	}
}

func benchFilterMapCountChain(b *testing.B, n int) {
	users := makeUsers(n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := From(users).KeepIf(benchUser.IsActive).ToString(benchUser.GetName).Len()
		_ = count
	}
}
