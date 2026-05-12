package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStateStoreRoundtrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TBA_UPLOADER_DATA_DIR", dir)

	s, err := openStateStore("2026mitt")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if s.snapshot().Config.EventKey != "2026mitt" {
		t.Fatalf("event key not set")
	}
	if err := s.update(func(es *eventState) {
		es.Config.ProfileName = "scratch"
		es.Config.PlaylistName = "Test"
		es.Videos["foo.mp4"] = &videoEntry{Size: 100, Status: statusStable}
		es.ManualVideoIDs["qm5"] = "abc11char23"
	}); err != nil {
		t.Fatalf("update: %v", err)
	}

	// Reopen and confirm persistence.
	s2, err := openStateStore("2026mitt")
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	got := s2.snapshot()
	if got.Config.ProfileName != "scratch" || got.Config.PlaylistName != "Test" {
		t.Errorf("config not persisted: %+v", got.Config)
	}
	if got.Videos["foo.mp4"].Size != 100 {
		t.Errorf("videos not persisted: %+v", got.Videos)
	}
	if got.ManualVideoIDs["qm5"] != "abc11char23" {
		t.Errorf("manual ids not persisted: %+v", got.ManualVideoIDs)
	}

	// File should exist at the expected path.
	want := filepath.Join(dir, "events", "2026mitt", "state.json")
	if _, err := os.Stat(want); err != nil {
		t.Errorf("state file missing: %v", err)
	}
}

func TestStateStoreSnapshotIsCopy(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TBA_UPLOADER_DATA_DIR", dir)
	s, _ := openStateStore("ev")
	_ = s.update(func(es *eventState) {
		es.Videos["a.mp4"] = &videoEntry{Status: statusStable}
	})
	snap := s.snapshot()
	snap.Videos["a.mp4"].Status = "tampered"
	// The internal state must not have been mutated by the caller's edit.
	if s.snapshot().Videos["a.mp4"].Status == "tampered" {
		t.Fatal("snapshot leaked a live reference")
	}
}
