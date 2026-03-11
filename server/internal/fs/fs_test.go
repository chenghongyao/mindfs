package fs

import "testing"

func TestRootInfoNormalizePathAcceptsAbsolutePathWithoutLeadingSlash(t *testing.T) {
	root := NewRootInfo("mindfs", "mindfs", "/Users/bixin/project/mindfs")

	got, err := root.NormalizePath("Users/bixin/project/mindfs/test.json")
	if err != nil {
		t.Fatalf("NormalizePath returned error: %v", err)
	}
	if got != "test.json" {
		t.Fatalf("NormalizePath = %q, want %q", got, "test.json")
	}
}

func TestRootInfoNormalizePathStripsFragment(t *testing.T) {
	root := NewRootInfo("mindfs", "mindfs", "/Users/bixin/project/mindfs")

	got, err := root.NormalizePath("Users/bixin/project/mindfs/design/test.md#L89")
	if err != nil {
		t.Fatalf("NormalizePath returned error: %v", err)
	}
	if got != "design/test.md" {
		t.Fatalf("NormalizePath = %q, want %q", got, "design/test.md")
	}
}
