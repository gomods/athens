package mongo

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestModuleDocumentBSONOmitsEmptyID(t *testing.T) {
	doc := moduleDocument{
		Module:  "github.com/gomods/athens",
		Version: "v1.2.3",
		Mod:     []byte("module x"),
		Info:    []byte(`{"Version":"v1.2.3"}`),
	}

	b, err := bson.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}

	var m bson.M
	if err := bson.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["_id"]; ok {
		t.Fatalf("expected zero ID to be omitted, got document %v", m)
	}
	if len(m) != 4 {
		t.Fatalf("expected exactly 4 fields, got %v", m)
	}
	if got := m["module"]; got != doc.Module {
		t.Errorf("module: got %v, want %v", got, doc.Module)
	}
	if got := m["version"]; got != doc.Version {
		t.Errorf("version: got %v, want %v", got, doc.Version)
	}
	for _, f := range []string{"mod", "info"} {
		if _, ok := m[f]; !ok {
			t.Errorf("missing field %q in %v", f, m)
		}
	}
}

func TestModuleDocumentBSONRoundTrip(t *testing.T) {
	want := moduleDocument{
		ID:      bson.NewObjectID(),
		Module:  "github.com/gomods/athens",
		Version: "v1.2.3",
		Mod:     []byte("module x"),
		Info:    []byte(`{"Version":"v1.2.3"}`),
	}

	b, err := bson.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}

	raw := bson.Raw(b)
	id, ok := raw.Lookup("_id").ObjectIDOK()
	if !ok || id != want.ID {
		t.Fatalf("_id: got %v (ok=%v), want %v", id, ok, want.ID)
	}

	var got moduleDocument
	if err := bson.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != want.ID || got.Module != want.Module || got.Version != want.Version ||
		string(got.Mod) != string(want.Mod) || string(got.Info) != string(want.Info) {
		t.Fatalf("round trip mismatch: got %+v, want %+v", got, want)
	}
}
