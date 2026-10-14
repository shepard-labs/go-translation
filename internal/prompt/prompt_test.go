package prompt

import (
	"strings"
	"testing"
)

func TestSystemAutoSource(t *testing.T) {
	got := System("auto", "Spanish")
	if !strings.Contains(got, "Detect the source language automatically") {
		t.Fatalf("expected auto-detect instruction, got %q", got)
	}
	if !strings.Contains(got, "Spanish") {
		t.Fatalf("expected target language, got %q", got)
	}
}

func TestUserWrapsContent(t *testing.T) {
	got := User("hello")
	if !strings.Contains(got, "<content>\nhello\n</content>") {
		t.Fatalf("expected wrapped content, got %q", got)
	}
}
