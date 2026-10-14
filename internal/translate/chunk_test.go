package translate

import (
	"strings"
	"testing"
)

func TestChunksReassemble(t *testing.T) {
	input := "first paragraph\n\nsecond paragraph\n\nthird paragraph"
	chunks := Chunks(input, 20)
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}
	if strings.Join(chunks, "") != input {
		t.Fatalf("chunks did not reassemble to original input")
	}
}

func TestChunksHardSplitsLongText(t *testing.T) {
	input := strings.Repeat("a", 25)
	chunks := Chunks(input, 10)
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
	if strings.Join(chunks, "") != input {
		t.Fatalf("chunks did not reassemble to original input")
	}
}
