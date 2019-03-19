package moniker

import (
	"testing"
)

func TestNamer(t *testing.T) {
	n := New().(*defaultNamer)
	n.Descriptor = []string{"foo"}
	n.Noun = []string{"bar"}

	if name := n.Name(); name != "foo bar" {
		t.Fatalf("Got %s", name)
	}

	if name := n.NameSep("$"); name != "foo$bar" {
		t.Fatalf("Got %s", name)
	}
}
