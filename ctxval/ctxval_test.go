package ctxval_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/binaryphile/fluentfp/ctxval"
)

type RequestID string
type TraceID string

type User struct{ Name string }

func TestWithAndFrom(t *testing.T) {
	ctx := ctxval.With(context.Background(), RequestID("abc"))
	got, ok := ctxval.Lookup[RequestID](ctx).Get()

	if !ok {
		t.Fatal("expected ok")
	}
	if got != "abc" {
		t.Fatalf("got %q, want %q", got, "abc")
	}
}

func TestFromAbsent(t *testing.T) {
	_, ok := ctxval.Lookup[RequestID](context.Background()).Get()

	if ok {
		t.Fatal("expected not-ok for absent key")
	}
}

func TestWithShadowsParent(t *testing.T) {
	parent := ctxval.With(context.Background(), RequestID("first"))
	child := ctxval.With(parent, RequestID("second"))

	got, ok := ctxval.Lookup[RequestID](child).Get()

	if !ok {
		t.Fatal("expected ok")
	}
	if got != "second" {
		t.Fatalf("got %q, want %q (child should shadow parent)", got, "second")
	}
}

func TestDistinctNamedTypes(t *testing.T) {
	ctx := ctxval.With(context.Background(), RequestID("req"))
	ctx = ctxval.With(ctx, TraceID("trace"))

	reqID, reqOK := ctxval.Lookup[RequestID](ctx).Get()
	trID, trOK := ctxval.Lookup[TraceID](ctx).Get()

	if !reqOK || reqID != "req" {
		t.Fatalf("RequestID: got %q ok=%v, want %q ok=true", reqID, reqOK, "req")
	}
	if !trOK || trID != "trace" {
		t.Fatalf("TraceID: got %q ok=%v, want %q ok=true", trID, trOK, "trace")
	}
}

func TestNilInterfaceValuePresent(t *testing.T) {
	// r is a true nil interface, not a typed-nil concrete.
	var r io.Reader

	ctx := ctxval.With(context.Background(), r)
	got, ok := ctxval.Lookup[io.Reader](ctx).Get()

	if !ok {
		t.Fatal("expected ok — nil interface should be present, not absent")
	}
	if got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

func TestInterfaceVsConcreteType(t *testing.T) {
	// reader is declared as io.Reader — static type determines the key.
	var reader io.Reader = &bytes.Buffer{}

	ctx := ctxval.With(context.Background(), reader)

	_, readerOK := ctxval.Lookup[io.Reader](ctx).Get()
	if !readerOK {
		t.Fatal("From[io.Reader] should find the value")
	}

	_, bufOK := ctxval.Lookup[*bytes.Buffer](ctx).Get()
	if bufOK {
		t.Fatal("From[*bytes.Buffer] should NOT find value stored as io.Reader")
	}
}

func TestValueVsPointerType(t *testing.T) {
	ctx := ctxval.With(context.Background(), User{Name: "alice"})

	_, valOK := ctxval.Lookup[User](ctx).Get()
	if !valOK {
		t.Fatal("From[User] should find the value")
	}

	_, ptrOK := ctxval.Lookup[*User](ctx).Get()
	if ptrOK {
		t.Fatal("From[*User] should NOT find value stored as User")
	}
}

func TestNonComparableTypeParameter(t *testing.T) {
	data := []byte("hello")

	ctx := ctxval.With[[]byte](context.Background(), data)
	got, ok := ctxval.Lookup[[]byte](ctx).Get()

	if !ok {
		t.Fatal("expected ok — []byte T should work (key struct is comparable)")
	}
	if string(got) != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}
}

func TestAnyTypeParameter(t *testing.T) {
	ctx := ctxval.With[any](context.Background(), "hello")
	got, ok := ctxval.Lookup[any](ctx).Get()

	if !ok {
		t.Fatal("expected ok")
	}
	if got != "hello" {
		t.Fatalf("got %v, want %q", got, "hello")
	}
}

func TestTypeAliasCollides(t *testing.T) {
	type Alias = string

	ctx := ctxval.With(context.Background(), Alias("via-alias"))
	got, ok := ctxval.Lookup[string](ctx).Get()

	if !ok {
		t.Fatal("expected ok — type alias shares key with underlying type")
	}
	if got != "via-alias" {
		t.Fatalf("got %q, want %q", got, "via-alias")
	}
}

func TestKeyWithAndFrom(t *testing.T) {
	key := ctxval.NewKey[string]()

	ctx := key.With(context.Background(), "value")
	got, ok := key.From(ctx).Get()

	if !ok {
		t.Fatal("expected ok")
	}
	if got != "value" {
		t.Fatalf("got %q, want %q", got, "value")
	}
}

func TestDistinctKeysForSameType(t *testing.T) {
	key1 := ctxval.NewKey[string]()
	key2 := ctxval.NewKey[string]()

	ctx := key1.With(context.Background(), "first")
	ctx = key2.With(ctx, "second")

	got1, _ := key1.From(ctx).Get()
	got2, _ := key2.From(ctx).Get()

	if got1 != "first" {
		t.Fatalf("key1: got %q, want %q", got1, "first")
	}
	if got2 != "second" {
		t.Fatalf("key2: got %q, want %q", got2, "second")
	}
}

func TestKeyNilInterfacePresent(t *testing.T) {
	key := ctxval.NewKey[io.Reader]()

	// r is a true nil interface.
	var r io.Reader

	ctx := key.With(context.Background(), r)
	got, ok := key.From(ctx).Get()

	if !ok {
		t.Fatal("expected ok — nil interface via Key should be present")
	}
	if got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

func TestKeyDoesNotCollideWithTypeKeyed(t *testing.T) {
	key := ctxval.NewKey[string]()

	ctx := ctxval.With(context.Background(), "type-keyed")
	ctx = key.With(ctx, "named-key")

	typeKeyed, _ := ctxval.Lookup[string](ctx).Get()
	namedKey, _ := key.From(ctx).Get()

	if typeKeyed != "type-keyed" {
		t.Fatalf("type-keyed: got %q, want %q", typeKeyed, "type-keyed")
	}
	if namedKey != "named-key" {
		t.Fatalf("named-key: got %q, want %q", namedKey, "named-key")
	}
}

func TestKeyShadowsParent(t *testing.T) {
	key := ctxval.NewKey[string]()

	parent := key.With(context.Background(), "first")
	child := key.With(parent, "second")

	got, ok := key.From(child).Get()

	if !ok {
		t.Fatal("expected ok")
	}
	if got != "second" {
		t.Fatalf("got %q, want %q (child should shadow parent)", got, "second")
	}
}

func TestKeyWrongTypeCharacterization(t *testing.T) {
	key := ctxval.NewKey[string]()

	// Context has no string value under this key.
	ctx := context.Background()

	_, ok := key.From(ctx).Get()
	if ok {
		t.Fatal("expected not-ok for absent key")
	}
}

func TestNilKeyPanics(t *testing.T) {
	t.Run("With", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()

		var k *ctxval.Key[string]

		k.With(context.Background(), "x")
	})

	t.Run("From", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()

		var k *ctxval.Key[string]

		k.From(context.Background())
	})
}

func TestWithNilContextPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for nil context")
		}
	}()

	ctxval.With[string](nil, "x")
}

func TestLookupNilContextPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for nil context")
		}
	}()

	ctxval.Lookup[string](nil)
}

func TestKeyWithNilContextPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for nil context")
		}
	}()

	key := ctxval.NewKey[string]()

	key.With(nil, "x")
}

func TestKeyFromNilContextPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for nil context")
		}
	}()

	key := ctxval.NewKey[string]()

	key.From(nil)
}
