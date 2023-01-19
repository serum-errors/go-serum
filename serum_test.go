package serum_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/serum-errors/go-serum"
)

func TestErrorsIs(t *testing.T) {
	nonSerumErr := fmt.Errorf("non-serum")
	nonSerumWrapped := fmt.Errorf("wrap: %w", nonSerumErr)
	t.Run("sanity check", func(t *testing.T) {
		if !errors.Is(nonSerumWrapped, nonSerumErr) {
			t.Fatal("fmt.Errorf wrapping should be fine")
		}
	})
	eqSynth := func(t *testing.T, a, b error, expectEqual bool) {
		ai, ok := a.(serum.ErrorInterface)
		if !ok {
			t.Fatalf("%T is not a a serum.ErrorInterface", a)
		}
		bi, ok := b.(serum.ErrorInterface)
		if !ok {
			t.Fatalf("%T is not a a serum.ErrorInterface", b)
		}
		as := serum.SynthesizeString(ai)
		bs := serum.SynthesizeString(bi)
		if (as == bs) != expectEqual {
			t.Fatalf("unexpected message synthesis:\n%s\n%s", as, bs)
		}
	}
	t.Run("equivalent serum errors", func(t *testing.T) {
		// should fail if !errors.Is
		t.Run("with only codes", func(t *testing.T) {
			a := serum.Error("test")
			b := serum.Error("test")
			if !errors.Is(a, b) {
				t.Fatal("errors should be equivalent")
			}
			eqSynth(t, a, b, true)
		})
		t.Run("with literal messages", func(t *testing.T) {
			a := serum.Error("test", serum.WithMessageLiteral("a thing!"))
			b := serum.Error("test", serum.WithMessageLiteral("a thing!"))
			if !errors.Is(a, b) {
				t.Fatal("errors should be equivalent")
			}
			eqSynth(t, a, b, true)
		})
		t.Run("with untemplated details", func(t *testing.T) {
			a := serum.Error("test", serum.WithMessageLiteral("a thing!"), serum.WithDetail("foo", "bar"))
			b := serum.Error("test", serum.WithMessageLiteral("a thing!"), serum.WithDetail("bar", "grill"))
			if !errors.Is(a, b) {
				t.Fatal("we don't care about details which don't synthesize into the message")
			}
			eqSynth(t, a, b, true)
		})
		t.Run("with matching templated details", func(t *testing.T) {
			a := serum.Error("test", serum.WithMessageLiteral("the thing is {{foo}}"), serum.WithDetail("foo", "bar"))
			b := serum.Error("test", serum.WithMessageLiteral("the thing is {{foo}}"), serum.WithDetail("foo", "bar"))
			if !errors.Is(a, b) {
				t.Fatal("errors that synthesize to equivalent messages should be equivalent")
			}
			eqSynth(t, a, b, true)
		})
		t.Run("with non-serum cause", func(t *testing.T) {
			withCause := serum.Error("test", serum.WithCause(nonSerumErr))
			if !errors.Is(withCause, nonSerumErr) {
				t.Fatal("unable to use errors.Is with a non-serum error")
			}
		})
		t.Run("with wrapped non-serum cause", func(t *testing.T) {
			withCauseWrapped := serum.Error("test", serum.WithCause(nonSerumWrapped))
			if !errors.Is(withCauseWrapped, nonSerumErr) {
				t.Fatal("unable to use errors.Is with a wrapped non-serum error")
			}
		})
		t.Run("with serum cause", func(t *testing.T) {
			err := serum.Error("test", serum.WithMessageLiteral("thing!"))
			wrapper := serum.Error("test-wrapper", serum.WithCause(err))
			if !errors.Is(wrapper, err) {
				t.Fatal("serum-cause wrapping should resolve")
			}
		})
	})
	t.Run("non-equivalent serum errors", func(t *testing.T) {
		t.Run("with different codes", func(t *testing.T) {
			a := serum.Error("test")
			b := serum.Error("other-test")
			if errors.Is(a, b) {
				t.Fatal("errors with different codes should not be considered equal")
			}
			eqSynth(t, a, b, false)
		})
		t.Run("with different messages", func(t *testing.T) {
			a := serum.Error("test", serum.WithMessageLiteral("foo"))
			b := serum.Error("test", serum.WithMessageLiteral("bar"))
			if errors.Is(a, b) {
				t.Fatal("errors with different codes should not be considered equal")
			}
			eqSynth(t, a, b, false)
		})
		t.Run("with non-matching templated details", func(t *testing.T) {
			a := serum.Error("test", serum.WithMessageTemplate("a thing: {{foo}}"), serum.WithDetail("foo", "bar"))
			b := serum.Error("test", serum.WithMessageTemplate("a thing: {{foo}}"), serum.WithDetail("foo", "grill"))
			if errors.Is(a, b) {
				t.Fatal("errors with templated messages that synthesize differently should not be considered equal")
			}
			eqSynth(t, a, b, false)
		})
	})
}
