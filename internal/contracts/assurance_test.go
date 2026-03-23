package contracts

import (
	"encoding/json"
	"testing"
	"time"
)

func TestComputeArtifactCanonicalJSONDeterministic(t *testing.T) {
	m1 := map[string]any{"b": "two", "a": "one"}
	m2 := map[string]any{"a": "one", "b": "two"}
	c1, err := ComputeArtifactCanonicalJSON(m1)
	if err != nil {
		t.Fatalf("compute canonical json m1: %v", err)
	}
	c2, err := ComputeArtifactCanonicalJSON(m2)
	if err != nil {
		t.Fatalf("compute canonical json m2: %v", err)
	}
	if c1 != c2 {
		t.Fatalf("canonical json mismatch: %q vs %q", c1, c2)
	}
}

func TestComputeSHA256Hex(t *testing.T) {
	ex := ComputeSHA256Hex([]byte("abc"))
	want := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if ex != want {
		t.Fatalf("sha256 mismatch: got %q want %q", ex, want)
	}
}

func TestGenerateReceiptFilenameUniqueness(t *testing.T) {
	t1 := time.Now().Unix()
	r1, err := GenerateReceiptID("context-pack", "subject-A", t1)
	if err != nil {
		t.Fatalf("generate receipt id r1: %v", err)
	}
	r2, err := GenerateReceiptID("context-pack", "subject-B", t1+1)
	if err != nil {
		t.Fatalf("generate receipt id r2: %v", err)
	}
	f1 := GenerateReceiptFilename(r1, ComputeSHA256Hex([]byte(r1)), "context-pack")
	f2 := GenerateReceiptFilename(r2, ComputeSHA256Hex([]byte(r2)), "context-pack")
	if f1 == f2 {
		t.Fatalf("expected different filenames for different receipts: %q", f1)
	}
}

func TestReceiptCanonicalizationFieldPopulated(t *testing.T) {
	r := ReceiptArtifact{
		AssuranceEnvelope: AssuranceEnvelope{
			SchemaVersion:    1,
			Kind:             "receipt",
			SubjectID:        "subject",
			CreatedAt:        42,
			Canonicalization: "runecontext-canonical-json-v1",
		},
		ReceiptID:   "rid",
		ReceiptHash: "rhash",
	}

	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal receipt artifact: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(b, &payload); err != nil {
		t.Fatalf("unmarshal receipt artifact json: %v", err)
	}

	if val, ok := payload["canonicalization"]; !ok {
		t.Fatalf("expected canonicalization key in json, got: %s", string(b))
	} else if val != "runecontext-canonical-json-v1" {
		t.Fatalf("expected canonicalization value, got %v", val)
	}
}

func TestGenerateReceiptIDReturnsNoError(t *testing.T) {
	id, err := GenerateReceiptID("context-pack", "subject-A", time.Now().Unix())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty id")
	}
}
