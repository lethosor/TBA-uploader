// Package ytstudio drives YouTube Studio through a persistent Chromium user
// profile via chromedp. It is the Go equivalent of pit-podcast's
// upload/youtube.py.
//
// The exported [Driver] is the interface the rest of the program talks to, so
// it can be swapped for a Playwright-backed implementation later without
// touching callers. The chromedp implementation lives in this package.
package ytstudio

import (
	"context"
	"errors"
	"time"
)

// ErrSessionExpired is returned when the persistent profile no longer holds
// a valid YouTube session and a re-authentication flow is required.
var ErrSessionExpired = errors.New("ytstudio: session expired, sign-in required")

// UploadInput is everything the driver needs to push one video.
type UploadInput struct {
	VideoPath     string
	Title         string
	Description   string
	ThumbnailPath string // empty => skip thumbnail step
	PlaylistName  string // exact playlist name; empty => skip playlist add
	Visibility    string // PUBLIC | UNLISTED | PRIVATE; empty defaults to PUBLIC
}

// UploadResult carries the YouTube video ID and the channel name observed
// during the run.
type UploadResult struct {
	VideoID     string
	ChannelName string
}

// Driver is the operation set the worker depends on. Real implementation is
// [ChromedpDriver]; tests can use a fake.
type Driver interface {
	// Upload performs an end-to-end upload. The provided context bounds the
	// entire run; callers should give it a long timeout (uploads of large
	// match recordings + YT processing easily reach 30+ minutes).
	Upload(ctx context.Context, profileName string, in UploadInput) (UploadResult, error)

	// CheckChannel opens YT Studio with the profile and returns the
	// currently-selected channel name, or ErrSessionExpired if the profile
	// has been signed out.
	CheckChannel(ctx context.Context, profileName string) (string, error)

	// Login spawns Brave non-headless and blocks until the operator closes
	// the window. The persistent profile is left in whatever state the
	// operator leaves it in.
	Login(ctx context.Context, profileName string) error
}

// Defaults applied when fields are zero. Exported so callers can tune.
var (
	DefaultUploadDeadline       = 90 * time.Minute
	DefaultChecksCompleteDeadline = 90 * time.Minute
	DefaultLoginDeadline        = 30 * time.Minute
)
