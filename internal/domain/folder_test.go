package domain

import (
	"testing"
	"time"
)

func TestFolderDefaults(t *testing.T) {
	f := Folder{}
	if f.ID != "" {
		t.Errorf("expected empty ID, got %q", f.ID)
	}
	if f.Name != "" {
		t.Errorf("expected empty Name, got %q", f.Name)
	}
}

func TestFolderFields(t *testing.T) {
	now := time.Now()
	f := Folder{
		ID:        "folder1",
		Name:      "Technology",
		CreatedAt: now,
	}
	if f.ID != "folder1" {
		t.Errorf("expected folder1, got %q", f.ID)
	}
	if f.Name != "Technology" {
		t.Errorf("expected Technology, got %q", f.Name)
	}
	if !f.CreatedAt.Equal(now) {
		t.Errorf("created_at mismatch")
	}
}
