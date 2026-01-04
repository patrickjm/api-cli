package secret

import "testing"

func TestMemoryStore(t *testing.T) {
	mem := NewMemoryStore()
	SetStore(mem)
	defer SetStore(nil)

	if err := Set("p", "default", "token", "abc"); err != nil {
		t.Fatalf("Set error: %v", err)
	}
	val, err := Get("p", "default", "token")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if val != "abc" {
		t.Fatalf("expected abc, got %q", val)
	}
	if err := Delete("p", "default", "token"); err != nil {
		t.Fatalf("Delete error: %v", err)
	}
	if _, err := Get("p", "default", "token"); err == nil {
		t.Fatalf("expected error for missing secret")
	}
}
