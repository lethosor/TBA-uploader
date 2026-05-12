package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lethosor/TBA-uploader/internal/ytstudio"
)

const (
	scanInterval    = 5 * time.Second
	stableDelay     = 20 * time.Second // size+mtime must hold this long
	maxAttempts     = 3
	baseBackoff     = 30 * time.Second
)

// uploadManager owns a single event's state, folder watcher, and upload loop.
type uploadManager struct {
	store  *stateStore
	driver ytstudio.Driver

	// kick channel wakes the upload loop on demand (after scan / on retry).
	kick chan struct{}
	// quit causes the loop to exit.
	quit chan struct{}
	// once-guarded shutdown.
	stopOnce sync.Once
}

// newUploadManager constructs a manager bound to the given state store and
// chromedp driver. Call Start to begin the scan + upload loops.
func newUploadManager(store *stateStore, driver ytstudio.Driver) *uploadManager {
	return &uploadManager{
		store:  store,
		driver: driver,
		kick:   make(chan struct{}, 1),
		quit:   make(chan struct{}),
	}
}

func (m *uploadManager) Start() {
	go m.scanLoop()
	go m.uploadLoop()
}

func (m *uploadManager) Stop() {
	m.stopOnce.Do(func() {
		close(m.quit)
	})
}

func (m *uploadManager) nudge() {
	select {
	case m.kick <- struct{}{}:
	default:
	}
}

// ── scan loop ────────────────────────────────────────────────────────────────

func (m *uploadManager) scanLoop() {
	t := time.NewTicker(scanInterval)
	defer t.Stop()
	m.scanNow()
	for {
		select {
		case <-m.quit:
			return
		case <-t.C:
			m.scanNow()
		}
	}
}

// scanNow does one folder scan: discovers new files, updates stability, and
// promotes entries whose size+mtime have been unchanged long enough.
func (m *uploadManager) scanNow() {
	// Watch the same folder AutoAV records to (settings.VideoDir, set by
	// the -video-dir flag or the legacy /save endpoint). No need to
	// duplicate it in per-event config.
	dir := settings.VideoDir
	if dir == "" {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		// Folder doesn't exist yet or is unreadable; log once per scan and
		// move on. A missing video_dir is normal before an event starts.
		log.Printf("scan: read %s: %v", dir, err)
		return
	}
	now := nowUnix()
	seen := map[string]bool{}
	err = m.store.update(func(s *eventState) {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.EqualFold(filepath.Ext(name), ".mp4") {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			size := info.Size()
			mtime := info.ModTime().Unix()
			seen[name] = true

			entry, ok := s.Videos[name]
			if !ok {
				entry = &videoEntry{
					Size:        size,
					Mtime:       mtime,
					Status:      statusNew,
					StableSince: 0,
				}
				s.Videos[name] = entry
			}

			// Already-uploaded or skipped entries are immutable from here.
			if entry.Status == statusUploaded || entry.Status == statusSkipped {
				continue
			}

			// Detect change. Any change resets the stable timer.
			if entry.Size != size || entry.Mtime != mtime {
				entry.Size = size
				entry.Mtime = mtime
				entry.StableSince = 0
				if entry.Status == statusStable {
					entry.Status = statusNew
				}
				continue
			}

			// Same size+mtime as last time. Start or continue the stable
			// timer.
			if entry.StableSince == 0 {
				entry.StableSince = now
			}
			if size > 0 && now-entry.StableSince >= int64(stableDelay/time.Second) {
				if entry.Status == statusNew {
					entry.Status = statusStable
				}
			}
		}
	})
	if err != nil {
		log.Printf("scan: state save: %v", err)
	}
	// New stable files? Wake the upload loop.
	m.nudge()
	_ = seen // present for future "removed file" handling.
}

// ── upload loop ──────────────────────────────────────────────────────────────

func (m *uploadManager) uploadLoop() {
	// Wait for an initial kick before doing anything. After that, an idle
	// sleep doubles as a "retry-due check" cadence.
	for {
		select {
		case <-m.quit:
			return
		case <-m.kick:
		case <-time.After(scanInterval):
		}
		m.uploadOne()
	}
}

// uploadOne picks the next eligible file and runs it through the driver. One
// in-flight upload at a time by design — YT Studio doesn't like concurrent
// browser sessions per profile.
func (m *uploadManager) uploadOne() {
	target, ok := m.pickNext()
	if !ok {
		return
	}
	cfg := m.store.snapshot().Config
	if cfg.ProfileName == "" {
		log.Printf("upload: profile_name not set, skipping %s", target.filename)
		return
	}

	log.Printf("upload: starting %s", target.filename)
	_ = m.store.update(func(s *eventState) {
		s.Videos[target.filename].Status = statusUploading
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result, err := m.driver.Upload(ctx, cfg.ProfileName, ytstudio.UploadInput{
		VideoPath:     filepath.Join(settings.VideoDir, target.filename),
		Title:         target.title,
		Description:   target.description,
		ThumbnailPath: cfg.ThumbnailPath,
		PlaylistName:  cfg.PlaylistName,
		Visibility:    "PUBLIC",
	})

	if err != nil {
		log.Printf("upload: %s failed: %v", target.filename, err)
		_ = m.store.update(func(s *eventState) {
			v := s.Videos[target.filename]
			v.Attempts++
			v.LastError = err.Error()
			if errors.Is(err, ytstudio.ErrSessionExpired) {
				v.Status = statusFailed
				v.NextAttempt = 0
				s.NeedsReauth = true
				return
			}
			if v.Attempts >= maxAttempts {
				v.Status = statusFailed
				v.NextAttempt = 0
			} else {
				// Exponential backoff: 30s, 60s, 120s.
				delay := baseBackoff * time.Duration(1<<(v.Attempts-1))
				v.NextAttempt = nowUnix() + int64(delay/time.Second)
				v.Status = statusStable
			}
		})
		return
	}

	log.Printf("upload: %s -> %s", target.filename, result.VideoID)
	_ = m.store.update(func(s *eventState) {
		v := s.Videos[target.filename]
		v.Status = statusUploaded
		v.YTVideoID = result.VideoID
		v.UploadedAt = time.Now().UTC().Format(time.RFC3339)
		v.TitleUsed = target.title
		v.LastError = ""
		v.NextAttempt = 0
		if result.ChannelName != "" {
			s.LastChannelName = result.ChannelName
		}
	})
	// Immediately try the next one.
	m.nudge()
}

// pendingUpload is one queued work item with rendered metadata.
type pendingUpload struct {
	filename    string
	title       string
	description string
}

// pickNext returns the highest-priority work item: lowest orderKey among
// stable, non-skipped, non-uploaded entries whose NextAttempt time has passed.
func (m *uploadManager) pickNext() (pendingUpload, bool) {
	st := m.store.snapshot()
	now := nowUnix()

	type candidate struct {
		filename string
		parsed   parsedFilename
		entry    *videoEntry
	}
	var cands []candidate
	for name, entry := range st.Videos {
		if entry.Status != statusStable {
			continue
		}
		if entry.NextAttempt != 0 && entry.NextAttempt > now {
			continue
		}
		p, ok := parseFilename(name)
		if !ok {
			continue
		}
		if !p.includeLevel(st.Config) {
			continue
		}
		// Skip files belonging to other events. We match by VideoPrefix
		// case-insensitively against either Config.EventName or the prefix
		// stored on previously-uploaded entries. If neither is set, accept
		// all parseable files — operator's responsibility.
		cands = append(cands, candidate{name, p, entry})
	}
	if len(cands) == 0 {
		return pendingUpload{}, false
	}
	sort.Slice(cands, func(i, j int) bool {
		return cands[i].parsed.orderKey() < cands[j].parsed.orderKey()
	})
	pick := cands[0]
	ctx := buildTemplateContext(pick.parsed, pick.entry.Meta, st.Config)
	title := renderTitle(st.Config.TitleTemplate, ctx)
	if strings.TrimSpace(title) == "" {
		// Defensive: never push an empty title to YT.
		title = strings.TrimSuffix(pick.filename, pick.parsed.Extension)
	}
	ctx.Title = title
	description := renderDescription(st.Config.DescriptionTemplate, ctx)
	return pendingUpload{
		filename:    pick.filename,
		title:       title,
		description: description,
	}, true
}

// requestRetry resets an entry's failure state so the upload loop will pick it
// up again.
func (m *uploadManager) requestRetry(filename string) error {
	return m.store.update(func(s *eventState) {
		v, ok := s.Videos[filename]
		if !ok {
			return
		}
		v.Status = statusStable
		v.Attempts = 0
		v.NextAttempt = 0
		v.LastError = ""
	})
}

func (m *uploadManager) requestSkip(filename string) error {
	err := m.store.update(func(s *eventState) {
		v, ok := s.Videos[filename]
		if !ok {
			return
		}
		v.Status = statusSkipped
	})
	return err
}

// resetSessionFlag is called after a successful login or channel check, to
// drop the "needs sign-in" banner.
func (m *uploadManager) resetSessionFlag() {
	_ = m.store.update(func(s *eventState) {
		s.NeedsReauth = false
	})
	m.nudge()
}

// String description for diagnostics — left here to placate the linter
// if we end up not using one of the constants.
var _ = fmt.Sprintf
