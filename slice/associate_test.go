package slice

import "testing"

func TestAssociate(t *testing.T) {
	type user struct {
		id   int
		name string
	}

	toEntry := func(u user) (int, string) { return u.id, u.name }

	t.Run("nil returns nil map", func(t *testing.T) {
		got := Associate[user](nil, toEntry)
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("empty returns nil map", func(t *testing.T) {
		got := Associate([]user{}, toEntry)
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("basic", func(t *testing.T) {
		users := []user{{1, "alice"}, {2, "bob"}}
		got := Associate(users, toEntry)
		if got[1] != "alice" || got[2] != "bob" || len(got) != 2 {
			t.Errorf("got %v", got)
		}
	})

	t.Run("duplicate keys last wins", func(t *testing.T) {
		users := []user{{1, "alice"}, {1, "bob"}}
		got := Associate(users, toEntry)
		if got[1] != "bob" || len(got) != 1 {
			t.Errorf("got %v, want map[1:bob]", got)
		}
	})
}
