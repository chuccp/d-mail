package rest

import "testing"

func TestFilePathOf(t *testing.T) {
	const files = `[{"name":"a.log","filePath":"/cache/a.log"},{"name":"b.pdf","filePath":"/cache/b.pdf"}]`

	t.Run("returns the stored path", func(t *testing.T) {
		got, err := FilePathOf(files, "b.pdf")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "/cache/b.pdf" {
			t.Errorf("got %q, want %q", got, "/cache/b.pdf")
		}
	})

	t.Run("empty files means no attachments", func(t *testing.T) {
		if _, err := FilePathOf("", "a.log"); err == nil {
			t.Error("expected an error for an empty files field")
		}
	})

	t.Run("invalid json is rejected", func(t *testing.T) {
		if _, err := FilePathOf("not json", "a.log"); err == nil {
			t.Error("expected an error for invalid JSON")
		}
	})

	t.Run("name not in the record is rejected", func(t *testing.T) {
		if _, err := FilePathOf(files, "/etc/passwd"); err == nil {
			t.Error("expected an error for a name that is not recorded")
		}
	})

	t.Run("entries without a path are skipped", func(t *testing.T) {
		if _, err := FilePathOf(`[{"name":"a.log"}]`, "a.log"); err == nil {
			t.Error("expected an error when the stored path is blank")
		}
	})

	t.Run("paths are matched by name only", func(t *testing.T) {
		// A caller must not be able to reach a path that is not recorded under its own name
		if _, err := FilePathOf(files, "a.log/../b.pdf"); err == nil {
			t.Error("expected an error for a traversal-style name")
		}
	})
}
