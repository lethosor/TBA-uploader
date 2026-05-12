package main

import (
	"regexp"
	"strconv"
	"strings"
)

// parsedFilename describes everything we can learn from a video filename
// produced by the autoAVRenameLastVideo flow in App.vue.
//
// Examples:
//
//	2026 Tornado Tumble Qualification Match 5.mp4
//	2026 Tornado Tumble Qualification Match 5 Play 2.mp4
//	2026 Tornado Tumble Final Match 1.mp4
//	2026 Tornado Tumble Qualification Match 5 - 2.mp4
type parsedFilename struct {
	VideoPrefix string // text before " {Level} Match ..." (e.g. "2026 Tornado Tumble")
	Level       string // "Test" | "Practice" | "Qualification" | "Playoff" | "Final" | "Manual"
	MatchNumber int
	Play        int    // 1 if "Play N" is absent
	Dedup       int    // 0 if no " - N" suffix
	Extension   string // e.g. ".mp4"
}

// parseFilename returns the parsed form and true on success, or zero+false if the
// filename doesn't match the expected shape. Order of suffixes matters: the
// "Play N" segment is parsed before the " - N" dedup tail.
var filenameRe = regexp.MustCompile(
	`^(.+?) (Test|Practice|Qualification|Playoff|Final|Manual) Match (\d+)(?: Play (\d+))?(?: - (\d+))?(\.[A-Za-z0-9]+)$`,
)

func parseFilename(name string) (parsedFilename, bool) {
	m := filenameRe.FindStringSubmatch(name)
	if m == nil {
		return parsedFilename{}, false
	}
	out := parsedFilename{
		VideoPrefix: m[1],
		Level:       m[2],
		Extension:   m[6],
		Play:        1,
	}
	out.MatchNumber, _ = strconv.Atoi(m[3])
	if m[4] != "" {
		out.Play, _ = strconv.Atoi(m[4])
	}
	if m[5] != "" {
		out.Dedup, _ = strconv.Atoi(m[5])
	}
	return out, true
}

// matchLabel is the human-readable label, e.g. "Qualification 5" or "Final 1".
func (p parsedFilename) matchLabel() string {
	return p.Level + " " + strconv.Itoa(p.MatchNumber)
}

// playSuffix renders " Play N" for replays, or empty for Play 1.
func (p parsedFilename) playSuffix() string {
	if p.Play <= 1 {
		return ""
	}
	return " Play " + strconv.Itoa(p.Play)
}

// includeLevel decides whether this filename's level passes the include_* config.
// Manual recordings are always skipped — they're ad-hoc recordings made
// outside the regular match flow and shouldn't be auto-uploaded.
func (p parsedFilename) includeLevel(cfg eventConfig) bool {
	switch strings.ToLower(p.Level) {
	case "qualification", "playoff", "final":
		return true
	case "practice":
		return cfg.IncludePractice
	case "test":
		return cfg.IncludeTest
	}
	return false
}

// orderKey returns a sortable integer for match order. Levels are bucketed so
// Quals come before Playoffs come before Finals. Within a level the match number
// (and then play number) takes over, so replays sort right after their parent.
func (p parsedFilename) orderKey() int64 {
	var bucket int64
	switch strings.ToLower(p.Level) {
	case "test":
		bucket = 0
	case "practice":
		bucket = 1
	case "qualification":
		bucket = 2
	case "playoff":
		bucket = 3
	case "final":
		bucket = 4
	case "manual":
		bucket = 5
	default:
		bucket = 9
	}
	// Pack: bucket | match_number | play | dedup. Plenty of headroom.
	return bucket*1_000_000_000 + int64(p.MatchNumber)*1_000_000 + int64(p.Play)*1_000 + int64(p.Dedup)
}
