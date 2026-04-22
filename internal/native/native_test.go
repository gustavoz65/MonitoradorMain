package native

import "testing"

func TestContainsBytes(t *testing.T) {
	if !ContainsBytes([]byte("hello world"), []byte("world")) {
		t.Fatal("expected true")
	}
	if ContainsBytes([]byte("hello"), []byte("world")) {
		t.Fatal("expected false")
	}
	if !ContainsBytes([]byte("abc"), []byte("")) {
		t.Fatal("empty pattern should match")
	}
}

func TestHashBytes(t *testing.T) {
	h1 := HashBytes([]byte("monimaster"))
	h2 := HashBytes([]byte("monimaster"))
	h3 := HashBytes([]byte("other"))
	if h1 != h2 {
		t.Fatal("same input must produce same hash")
	}
	if h1 == h3 {
		t.Fatal("different inputs should produce different hashes")
	}
}
