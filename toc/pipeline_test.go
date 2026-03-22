package toc_test

import (
	"strings"
	"testing"

	"github.com/binaryphile/fluentfp/toc"
)

// dummyStats returns a no-op stats function for pipeline registration.
func dummyStats() func() toc.Stats {
	return func() toc.Stats { return toc.Stats{} }
}

// linearPipeline builds A → B → C and freezes.
func linearPipeline() *toc.Pipeline {
	p := toc.NewPipeline()
	p.AddStage("A", dummyStats())
	p.AddStage("B", dummyStats())
	p.AddStage("C", dummyStats())
	p.AddEdge("A", "B")
	p.AddEdge("B", "C")
	p.Freeze()
	return p
}

// diamondPipeline builds A → {B, C} → D and freezes.
func diamondPipeline() *toc.Pipeline {
	p := toc.NewPipeline()
	p.AddStage("A", dummyStats())
	p.AddStage("B", dummyStats())
	p.AddStage("C", dummyStats())
	p.AddStage("D", dummyStats())
	p.AddEdge("A", "B")
	p.AddEdge("A", "C")
	p.AddEdge("B", "D")
	p.AddEdge("C", "D")
	p.Freeze()
	return p
}

// multiHeadPipeline builds two independent sources merging:
//
//	X → M
//	Y → M → Z
func multiHeadPipeline() *toc.Pipeline {
	p := toc.NewPipeline()
	p.AddStage("X", dummyStats())
	p.AddStage("Y", dummyStats())
	p.AddStage("M", dummyStats())
	p.AddStage("Z", dummyStats())
	p.AddEdge("X", "M")
	p.AddEdge("Y", "M")
	p.AddEdge("M", "Z")
	p.Freeze()
	return p
}

func TestAddStagePanics(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		wantMsg string
	}{
		{
			name: "empty_name",
			setup: func() {
				p := toc.NewPipeline()
				p.AddStage("", dummyStats())
			},
			wantMsg: "name must not be empty",
		},
		{
			name: "nil_stats",
			setup: func() {
				p := toc.NewPipeline()
				p.AddStage("A", nil)
			},
			wantMsg: "stats must not be nil",
		},
		{
			name: "duplicate",
			setup: func() {
				p := toc.NewPipeline()
				p.AddStage("A", dummyStats())
				p.AddStage("A", dummyStats())
			},
			wantMsg: "duplicate stage",
		},
		{
			name: "after_freeze",
			setup: func() {
				p := toc.NewPipeline()
				p.AddStage("A", dummyStats())
				p.Freeze()
				p.AddStage("B", dummyStats())
			},
			wantMsg: "already frozen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatal("expected panic")
				}
				msg, ok := r.(string)
				if !ok {
					t.Fatalf("panic value not string: %v", r)
				}
				if !strings.Contains(msg, tt.wantMsg) {
					t.Errorf("panic = %q, want substring %q", msg, tt.wantMsg)
				}
			}()
			tt.setup()
		})
	}
}

func TestAddEdgePanics(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		wantMsg string
	}{
		{
			name: "unknown_from",
			setup: func() {
				p := toc.NewPipeline()
				p.AddStage("B", dummyStats())
				p.AddEdge("A", "B")
			},
			wantMsg: "unknown stage in edge 'from'",
		},
		{
			name: "unknown_to",
			setup: func() {
				p := toc.NewPipeline()
				p.AddStage("A", dummyStats())
				p.AddEdge("A", "B")
			},
			wantMsg: "unknown stage in edge 'to'",
		},
		{
			name: "self_loop",
			setup: func() {
				p := toc.NewPipeline()
				p.AddStage("A", dummyStats())
				p.AddEdge("A", "A")
			},
			wantMsg: "self-loop",
		},
		{
			name: "duplicate_edge",
			setup: func() {
				p := toc.NewPipeline()
				p.AddStage("A", dummyStats())
				p.AddStage("B", dummyStats())
				p.AddEdge("A", "B")
				p.AddEdge("A", "B")
			},
			wantMsg: "duplicate edge",
		},
		{
			name: "after_freeze",
			setup: func() {
				p := toc.NewPipeline()
				p.AddStage("A", dummyStats())
				p.AddStage("B", dummyStats())
				p.AddEdge("A", "B")
				p.Freeze()
				p.AddEdge("A", "B")
			},
			wantMsg: "already frozen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatal("expected panic")
				}
				msg, ok := r.(string)
				if !ok {
					t.Fatalf("panic value not string: %v", r)
				}
				if !strings.Contains(msg, tt.wantMsg) {
					t.Errorf("panic = %q, want substring %q", msg, tt.wantMsg)
				}
			}()
			tt.setup()
		})
	}
}

func TestFreezePanics(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		wantMsg string
	}{
		{
			name: "empty",
			setup: func() {
				p := toc.NewPipeline()
				p.Freeze()
			},
			wantMsg: "no stages",
		},
		{
			name: "cycle",
			setup: func() {
				p := toc.NewPipeline()
				p.AddStage("A", dummyStats())
				p.AddStage("B", dummyStats())
				p.AddStage("C", dummyStats())
				p.AddEdge("A", "B")
				p.AddEdge("B", "C")
				p.AddEdge("C", "A")
				p.Freeze()
			},
			wantMsg: "cycle",
		},
		{
			name: "double_freeze",
			setup: func() {
				p := toc.NewPipeline()
				p.AddStage("A", dummyStats())
				p.Freeze()
				p.Freeze()
			},
			wantMsg: "already frozen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatal("expected panic")
				}
				msg, ok := r.(string)
				if !ok {
					t.Fatalf("panic value not string: %v", r)
				}
				if !strings.Contains(msg, tt.wantMsg) {
					t.Errorf("panic = %q, want substring %q", msg, tt.wantMsg)
				}
			}()
			tt.setup()
		})
	}
}

func TestStages(t *testing.T) {
	p := linearPipeline()
	got := p.Stages()
	want := []string{"A", "B", "C"}

	if len(got) != len(want) {
		t.Fatalf("Stages() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Stages()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestHeads(t *testing.T) {
	t.Run("linear", func(t *testing.T) {
		p := linearPipeline()
		got := p.Heads()
		if len(got) != 1 || got[0] != "A" {
			t.Errorf("Heads() = %v, want [A]", got)
		}
	})

	t.Run("diamond", func(t *testing.T) {
		p := diamondPipeline()
		got := p.Heads()
		if len(got) != 1 || got[0] != "A" {
			t.Errorf("Heads() = %v, want [A]", got)
		}
	})

	t.Run("multi_head", func(t *testing.T) {
		p := multiHeadPipeline()
		got := p.Heads()
		if len(got) != 2 {
			t.Fatalf("Heads() = %v, want 2 heads", got)
		}
		if got[0] != "X" || got[1] != "Y" {
			t.Errorf("Heads() = %v, want [X Y]", got)
		}
	})

	t.Run("single_stage", func(t *testing.T) {
		p := toc.NewPipeline()
		p.AddStage("only", dummyStats())
		p.Freeze()
		got := p.Heads()
		if len(got) != 1 || got[0] != "only" {
			t.Errorf("Heads() = %v, want [only]", got)
		}
	})
}

func TestHeadsTo(t *testing.T) {
	t.Run("multi_head_both_reach", func(t *testing.T) {
		p := multiHeadPipeline()
		got := p.HeadsTo("Z")
		if len(got) != 2 {
			t.Fatalf("HeadsTo(Z) = %v, want 2", got)
		}
	})

	t.Run("multi_head_one_reaches", func(t *testing.T) {
		// X → M → Z, Y → N (independent)
		p := toc.NewPipeline()
		p.AddStage("X", dummyStats())
		p.AddStage("Y", dummyStats())
		p.AddStage("M", dummyStats())
		p.AddStage("N", dummyStats())
		p.AddStage("Z", dummyStats())
		p.AddEdge("X", "M")
		p.AddEdge("M", "Z")
		p.AddEdge("Y", "N")
		p.Freeze()

		got := p.HeadsTo("Z")
		if len(got) != 1 || got[0] != "X" {
			t.Errorf("HeadsTo(Z) = %v, want [X]", got)
		}
	})

	t.Run("head_is_target", func(t *testing.T) {
		p := linearPipeline()
		got := p.HeadsTo("A")
		if len(got) != 1 || got[0] != "A" {
			t.Errorf("HeadsTo(A) = %v, want [A]", got)
		}
	})
}

func TestAncestorsOf(t *testing.T) {
	t.Run("linear_tail", func(t *testing.T) {
		p := linearPipeline()
		got := p.AncestorsOf("C")
		// BFS: B first (direct predecessor), then A.
		if len(got) != 2 || got[0] != "B" || got[1] != "A" {
			t.Errorf("AncestorsOf(C) = %v, want [B A]", got)
		}
	})

	t.Run("linear_middle", func(t *testing.T) {
		p := linearPipeline()
		got := p.AncestorsOf("B")
		if len(got) != 1 || got[0] != "A" {
			t.Errorf("AncestorsOf(B) = %v, want [A]", got)
		}
	})

	t.Run("linear_head", func(t *testing.T) {
		p := linearPipeline()
		got := p.AncestorsOf("A")
		if len(got) != 0 {
			t.Errorf("AncestorsOf(A) = %v, want []", got)
		}
	})

	t.Run("diamond", func(t *testing.T) {
		p := diamondPipeline()
		got := p.AncestorsOf("D")
		// Direct predecessors B, C (registration order), then A.
		if len(got) != 3 {
			t.Fatalf("AncestorsOf(D) = %v, want 3 ancestors", got)
		}
		// B and C are both direct predecessors; order depends on edge registration.
		if got[2] != "A" {
			t.Errorf("AncestorsOf(D)[2] = %q, want A", got[2])
		}
		// B and C should be in positions 0-1.
		bc := map[string]bool{got[0]: true, got[1]: true}
		if !bc["B"] || !bc["C"] {
			t.Errorf("AncestorsOf(D)[:2] = %v, want {B, C}", got[:2])
		}
	})

	t.Run("single_stage", func(t *testing.T) {
		p := toc.NewPipeline()
		p.AddStage("only", dummyStats())
		p.Freeze()
		got := p.AncestorsOf("only")
		if len(got) != 0 {
			t.Errorf("AncestorsOf(only) = %v, want []", got)
		}
	})
}

func TestDirectPredecessors(t *testing.T) {
	t.Run("linear", func(t *testing.T) {
		p := linearPipeline()
		got := p.DirectPredecessors("C")
		if len(got) != 1 || got[0] != "B" {
			t.Errorf("DirectPredecessors(C) = %v, want [B]", got)
		}
	})

	t.Run("diamond_merge", func(t *testing.T) {
		p := diamondPipeline()
		got := p.DirectPredecessors("D")
		if len(got) != 2 {
			t.Fatalf("DirectPredecessors(D) = %v, want 2", got)
		}
		preds := map[string]bool{got[0]: true, got[1]: true}
		if !preds["B"] || !preds["C"] {
			t.Errorf("DirectPredecessors(D) = %v, want {B, C}", got)
		}
	})

	t.Run("head_has_none", func(t *testing.T) {
		p := linearPipeline()
		got := p.DirectPredecessors("A")
		if got != nil {
			t.Errorf("DirectPredecessors(A) = %v, want nil", got)
		}
	})
}

func TestHasPath(t *testing.T) {
	tests := []struct {
		name string
		p    *toc.Pipeline
		from string
		to   string
		want bool
	}{
		{"linear_A_to_C", linearPipeline(), "A", "C", true},
		{"linear_C_to_A", linearPipeline(), "C", "A", false},
		{"linear_self", linearPipeline(), "A", "A", true},
		{"diamond_A_to_D", diamondPipeline(), "A", "D", true},
		{"diamond_B_to_C", diamondPipeline(), "B", "C", false},
		{"multi_X_to_Z", multiHeadPipeline(), "X", "Z", true},
		{"multi_Y_to_Z", multiHeadPipeline(), "Y", "Z", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.HasPath(tt.from, tt.to)
			if got != tt.want {
				t.Errorf("HasPath(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestIncoming(t *testing.T) {
	t.Run("linear", func(t *testing.T) {
		p := linearPipeline()
		got := p.Incoming("C")
		if len(got) != 1 {
			t.Fatalf("Incoming(C) = %v, want 1 edge", got)
		}
		if got[0].From != "B" || got[0].Ratio != 1 {
			t.Errorf("Incoming(C)[0] = {%q, %d}, want {B, 1}", got[0].From, got[0].Ratio)
		}
	})

	t.Run("with_ratios", func(t *testing.T) {
		p := toc.NewPipeline()
		p.AddStage("src_a", dummyStats())
		p.AddStage("src_b", dummyStats())
		p.AddStage("merge", dummyStats())
		p.AddEdgeWithRatio("src_a", "merge", 2)
		p.AddEdgeWithRatio("src_b", "merge", 1)
		p.Freeze()

		got := p.Incoming("merge")
		if len(got) != 2 {
			t.Fatalf("Incoming(merge) = %v, want 2 edges", got)
		}
		ratios := make(map[string]int)
		for _, e := range got {
			ratios[e.From] = e.Ratio
		}
		if ratios["src_a"] != 2 {
			t.Errorf("src_a ratio = %d, want 2", ratios["src_a"])
		}
		if ratios["src_b"] != 1 {
			t.Errorf("src_b ratio = %d, want 1", ratios["src_b"])
		}
	})

	t.Run("head_has_none", func(t *testing.T) {
		p := linearPipeline()
		got := p.Incoming("A")
		if got != nil {
			t.Errorf("Incoming(A) = %v, want nil", got)
		}
	})
}

func TestEdgeRatio(t *testing.T) {
	t.Run("default_ratio", func(t *testing.T) {
		p := linearPipeline()
		if got := p.EdgeRatio("A", "B"); got != 1 {
			t.Errorf("EdgeRatio(A,B) = %d, want 1", got)
		}
	})

	t.Run("explicit_ratio", func(t *testing.T) {
		p := toc.NewPipeline()
		p.AddStage("src_a", dummyStats())
		p.AddStage("src_b", dummyStats())
		p.AddStage("merge", dummyStats())
		p.AddEdgeWithRatio("src_a", "merge", 2) // 2 items from A per output
		p.AddEdgeWithRatio("src_b", "merge", 1) // 1 item from B per output
		p.Freeze()

		if got := p.EdgeRatio("src_a", "merge"); got != 2 {
			t.Errorf("EdgeRatio(src_a,merge) = %d, want 2", got)
		}
		if got := p.EdgeRatio("src_b", "merge"); got != 1 {
			t.Errorf("EdgeRatio(src_b,merge) = %d, want 1", got)
		}
	})

	t.Run("zero_ratio_panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for zero ratio")
			}
		}()
		p := toc.NewPipeline()
		p.AddStage("A", dummyStats())
		p.AddStage("B", dummyStats())
		p.AddEdgeWithRatio("A", "B", 0)
	})

	t.Run("unknown_edge_panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for unknown edge")
			}
		}()
		p := linearPipeline()
		p.EdgeRatio("A", "C") // no direct edge A→C
	})
}

func TestStageStats(t *testing.T) {
	called := false
	stats := func() toc.Stats {
		called = true
		return toc.Stats{Submitted: 42}
	}

	p := toc.NewPipeline()
	p.AddStage("A", stats)
	p.Freeze()

	fn := p.StageStats("A")
	s := fn()
	if !called {
		t.Error("stats function not called")
	}
	if s.Submitted != 42 {
		t.Errorf("Submitted = %d, want 42", s.Submitted)
	}
}

func TestNotFrozenPanics(t *testing.T) {
	methods := []struct {
		name string
		call func(*toc.Pipeline)
	}{
		{"Stages", func(p *toc.Pipeline) { p.Stages() }},
		{"Heads", func(p *toc.Pipeline) { p.Heads() }},
		{"HeadsTo", func(p *toc.Pipeline) { p.HeadsTo("A") }},
		{"AncestorsOf", func(p *toc.Pipeline) { p.AncestorsOf("A") }},
		{"DirectPredecessors", func(p *toc.Pipeline) { p.DirectPredecessors("A") }},
		{"HasPath", func(p *toc.Pipeline) { p.HasPath("A", "A") }},
		{"StageStats", func(p *toc.Pipeline) { p.StageStats("A") }},
	}

	for _, m := range methods {
		t.Run(m.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatal("expected panic for unfrozen pipeline")
				}
			}()
			p := toc.NewPipeline()
			p.AddStage("A", dummyStats())
			m.call(p)
		})
	}
}

