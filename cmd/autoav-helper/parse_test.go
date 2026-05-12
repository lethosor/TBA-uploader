package main

import "testing"

func TestParseFilename(t *testing.T) {
	cases := []struct {
		in       string
		ok       bool
		level    string
		num      int
		play     int
		dedup    int
		prefix   string
		ext      string
		label    string
		playSfx  string
		ordering int64
	}{
		{
			in: "2026 Tornado Tumble Qualification Match 5.mp4",
			ok: true, level: "Qualification", num: 5, play: 1, dedup: 0,
			prefix: "2026 Tornado Tumble", ext: ".mp4",
			label: "Qualification 5", playSfx: "",
		},
		{
			in: "2026 Tornado Tumble Qualification Match 5 Play 2.mp4",
			ok: true, level: "Qualification", num: 5, play: 2, dedup: 0,
			prefix: "2026 Tornado Tumble", ext: ".mp4",
			label: "Qualification 5", playSfx: " Play 2",
		},
		{
			in: "2026 Tornado Tumble Final Match 1.mp4",
			ok: true, level: "Final", num: 1, play: 1, dedup: 0,
			prefix: "2026 Tornado Tumble", ext: ".mp4",
			label: "Final 1", playSfx: "",
		},
		{
			in: "2026 Tornado Tumble Qualification Match 5 - 2.mp4",
			ok: true, level: "Qualification", num: 5, play: 1, dedup: 2,
			prefix: "2026 Tornado Tumble", ext: ".mp4",
		},
		{
			in: "2026 Tornado Tumble Practice Match 3.mp4",
			ok: true, level: "Practice", num: 3, play: 1,
		},
		{in: "garbage.mp4", ok: false},
		{in: "Random vmix recording.mp4", ok: false},
	}
	for _, c := range cases {
		got, ok := parseFilename(c.in)
		if ok != c.ok {
			t.Errorf("%q: ok=%v want %v", c.in, ok, c.ok)
			continue
		}
		if !ok {
			continue
		}
		if got.Level != c.level || got.MatchNumber != c.num || got.Play != c.play || got.Dedup != c.dedup {
			t.Errorf("%q: got %+v", c.in, got)
		}
		if c.prefix != "" && got.VideoPrefix != c.prefix {
			t.Errorf("%q: prefix %q want %q", c.in, got.VideoPrefix, c.prefix)
		}
		if c.label != "" && got.matchLabel() != c.label {
			t.Errorf("%q: label %q want %q", c.in, got.matchLabel(), c.label)
		}
		if c.ext != "" && got.Extension != c.ext {
			t.Errorf("%q: ext %q want %q", c.in, got.Extension, c.ext)
		}
		if c.playSfx != got.playSuffix() {
			t.Errorf("%q: playSuffix %q want %q", c.in, got.playSuffix(), c.playSfx)
		}
	}
}

func TestOrderKey(t *testing.T) {
	// Match-order invariants the worker depends on:
	//   - Quals before Playoffs before Finals
	//   - Within a level, ascending match number
	//   - Replays right after their parent (same match number, higher play)
	q1, _ := parseFilename("2026 X Qualification Match 1.mp4")
	q2, _ := parseFilename("2026 X Qualification Match 2.mp4")
	q2p2, _ := parseFilename("2026 X Qualification Match 2 Play 2.mp4")
	p1, _ := parseFilename("2026 X Playoff Match 1.mp4")
	f1, _ := parseFilename("2026 X Final Match 1.mp4")

	order := []parsedFilename{q1, q2, q2p2, p1, f1}
	for i := 1; i < len(order); i++ {
		if order[i-1].orderKey() >= order[i].orderKey() {
			t.Fatalf("order %d not before %d: %d vs %d",
				i-1, i, order[i-1].orderKey(), order[i].orderKey())
		}
	}
}

func TestIncludeLevel(t *testing.T) {
	cfg := eventConfig{IncludePractice: false, IncludeTest: false}
	p, _ := parseFilename("2026 X Qualification Match 5.mp4")
	if !p.includeLevel(cfg) {
		t.Fatal("qual should always be included")
	}
	prac, _ := parseFilename("2026 X Practice Match 5.mp4")
	if prac.includeLevel(cfg) {
		t.Fatal("practice should be excluded when IncludePractice=false")
	}
	cfg.IncludePractice = true
	if !prac.includeLevel(cfg) {
		t.Fatal("practice should be included when IncludePractice=true")
	}
	man, _ := parseFilename("2026 X Manual Match 1.mp4")
	if man.includeLevel(cfg) {
		t.Fatal("manual recordings should always be excluded")
	}
}
