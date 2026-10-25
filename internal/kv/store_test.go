package kv

import "testing"

func TestSetAndGet(t *testing.T) {
	store := NewStore()
	store.Set("name", "Riya")
	value, exists := store.Get("name")
	if !exists {
		t.Fatal("expected key to exist")
	}

	if value != "Riya" {
		t.Fatalf("expected Riya, got %s", value)
	}
}

func TestDel(t *testing.T) {
	store := NewStore()
	store.Set("name", "Riya")
	store.Delete("name")

	_, exists := store.Get("name")

	if exists {
		t.Fatal("expected key to be deleted")
	}
}
