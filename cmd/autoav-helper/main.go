package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/lethosor/TBA-uploader/internal/ytstudio"
)

var settings struct {
	VideoDir string
}

type FileInfo struct {
	Name  string `json:"name"`
	Mtime int64  `json:"mtime"`
}

// Module-level state populated in main().
var (
	managers   = map[string]*uploadManager{}
	managersMu sync.Mutex
	driver     ytstudio.Driver
)

func main() {
	addr := flag.String("listen", ":8807", "address to listen on")
	browserExe := flag.String("browser", "", "explicit browser executable path (else autodetect)")
	flag.StringVar(&settings.VideoDir, "video-dir", "/tmp/videos", "default folder containing recorded videos")
	flag.Parse()

	// Driver shared by all event managers. ProfileRoot is the global
	// {data_root}/profiles directory.
	d := ytstudio.NewChromedpDriver(
		filepath.Join(dataRoot(), "profiles"),
		*browserExe,
	)
	d.Verbose = true
	driver = d

	lock := sync.Mutex{}
	mux := http.NewServeMux()
	handle := func(method string, p string, handler func(w http.ResponseWriter, r *http.Request)) {
		mux.HandleFunc(p, func(w http.ResponseWriter, r *http.Request) {
			lock.Lock()
			defer lock.Unlock()
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Internal error: %v\n%s", err, debug.Stack())
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(fmt.Sprintf("Internal error: %v", err)))
				}
			}()

			log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
			w.Header().Set("access-control-allow-origin", "*")
			w.Header().Set("access-control-allow-methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("access-control-allow-headers", "content-type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if method != "" && method != r.Method {
				w.WriteHeader(http.StatusMethodNotAllowed)
				_, _ = w.Write([]byte("method not allowed: " + r.Method))
				return
			}
			handler(w, r)
		})
	}

	// Existing endpoints.
	handle(http.MethodGet, "/", handleRoot)
	handle(http.MethodPost, "/save", handleSaveSettings)
	handle(http.MethodGet, "/api/list", apiList)
	// /api/rename accepts both GET (legacy query params) and POST (JSON with optional meta).
	handle("", "/api/rename", apiRename)

	// YouTube auto-upload endpoints. /api/upload/config dispatches on
	// method (GET to read, POST to save) so we register it once.
	handle("", "/api/upload/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			apiUploadGetConfig(w, r)
		case http.MethodPost:
			apiUploadSaveConfig(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	handle(http.MethodPost, "/api/upload/thumbnail", apiUploadThumbnail)
	handle(http.MethodGet, "/api/upload/state", apiUploadGetState)
	handle(http.MethodPost, "/api/upload/scan", apiUploadScan)
	handle(http.MethodPost, "/api/upload/retry", apiUploadRetry)
	handle(http.MethodPost, "/api/upload/skip", apiUploadSkip)
	handle(http.MethodGet, "/api/upload/profiles", apiUploadListProfiles)
	handle(http.MethodPost, "/api/upload/profile/login", apiUploadProfileLogin)
	handle(http.MethodGet, "/api/upload/profile/check", apiUploadProfileCheck)
	// POST sets, DELETE clears.
	handle("", "/api/videos/manual", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			apiVideosManualSet(w, r)
		case http.MethodDelete:
			apiVideosManualDelete(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	log.Printf("listening on %s", *addr)
	_ = http.ListenAndServe(*addr, mux)
}

// ── helpers ────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("content-type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func requireEventKey(r *http.Request, w http.ResponseWriter) (string, bool) {
	key := r.URL.Query().Get("event_key")
	if key == "" {
		writeJSONError(w, http.StatusBadRequest, "event_key query parameter is required")
		return "", false
	}
	return key, true
}

// getOrCreateManager returns the long-lived uploadManager for the given event,
// starting its background loops on first access.
func getOrCreateManager(eventKey string) (*uploadManager, error) {
	managersMu.Lock()
	defer managersMu.Unlock()
	if m, ok := managers[eventKey]; ok {
		return m, nil
	}
	store, err := openStateStore(eventKey)
	if err != nil {
		return nil, err
	}
	m := newUploadManager(store, driver)
	managers[eventKey] = m
	m.Start()
	return m, nil
}

// ── legacy "settings" endpoints (Vue calls /api/list and /api/rename) ─────

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html")
	_, _ = fmt.Fprintf(w, `
		<form action="/save" method="POST">
			<label>
				Video folder:
				<input name="VideoDir" value="%s">
			</label>
			<br><br>
			<input type="submit" value="Save">
		</form>
	`, html.EscapeString(settings.VideoDir))
}

func handleSaveSettings(w http.ResponseWriter, r *http.Request) {
	if val := r.FormValue("VideoDir"); val != "" {
		settings.VideoDir = val
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func apiList(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	suffix := r.URL.Query().Get("suffix")

	allFiles, err := os.ReadDir(settings.VideoDir)
	if err != nil {
		panic(err)
	}

	files := make([]FileInfo, 0)
	for _, file := range allFiles {
		if !file.Type().IsRegular() {
			continue
		}
		if !strings.HasPrefix(file.Name(), prefix) || !strings.HasSuffix(file.Name(), suffix) {
			continue
		}
		info, err := file.Info()
		if err != nil {
			continue
		}
		files = append(files, FileInfo{
			Name:  file.Name(),
			Mtime: info.ModTime().Unix(),
		})
	}

	out, err := json.Marshal(files)
	if err != nil {
		panic(err)
	}
	_, _ = w.Write(out)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return !os.IsNotExist(err)
}

// apiRename handles both the legacy GET form (old_name, new_name query) and
// the new POST form (JSON body with optional event_key + meta).
//
// When event_key + meta are supplied, we also record the meta on the matching
// videoEntry in the per-event state so the upload pipeline can use it.
func apiRename(w http.ResponseWriter, r *http.Request) {
	var (
		oldName  string
		newName  string
		eventKey string
		meta     *videoMeta
	)

	if r.Method == http.MethodPost {
		var body struct {
			OldName  string     `json:"old_name"`
			NewName  string     `json:"new_name"`
			EventKey string     `json:"event_key"`
			Meta     *videoMeta `json:"meta,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		oldName = body.OldName
		newName = body.NewName
		eventKey = body.EventKey
		meta = body.Meta
	} else {
		oldName = r.URL.Query().Get("old_name")
		newName = r.URL.Query().Get("new_name")
		eventKey = r.URL.Query().Get("event_key")
	}

	if oldName == "" || newName == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("old_name and new_name are required"))
		return
	}

	oldPath := path.Join(settings.VideoDir, oldName)
	newPath := path.Join(settings.VideoDir, newName)

	if !fileExists(oldPath) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("old_path not found: " + oldPath))
		return
	}

	if fileExists(newPath) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte("new_path already exists: " + newPath))
		return
	}

	for i := 1; i <= 5; i++ {
		if i > 1 {
			log.Printf("retrying %d/5...", i)
		}
		if err := os.Rename(oldPath, newPath); err != nil {
			panic(err)
		}
		// For GDrive: rename can sometimes report success but actually fail.
		// Wait for a bit before checking whether the expected change actually
		// took place.
		time.Sleep(1 * time.Second)
		if fileExists(newPath) && !fileExists(oldPath) {
			break
		}
	}

	// Persist the meta on the matching state entry so the upload pipeline
	// can use it. Done after the rename so we know the file lives at the
	// new name.
	if meta != nil && eventKey != "" {
		m, err := getOrCreateManager(eventKey)
		if err != nil {
			log.Printf("apiRename: get manager for %s: %v", eventKey, err)
		} else {
			err := m.store.update(func(s *eventState) {
				entry, ok := s.Videos[newName]
				if !ok {
					entry = &videoEntry{Status: statusNew}
					s.Videos[newName] = entry
				}
				entry.Meta = meta
			})
			if err != nil {
				log.Printf("apiRename: persist meta: %v", err)
			}
			m.nudge()
		}
	}

	_, _ = w.Write([]byte("{\"ok\": true}"))
}

// ── /api/upload/* endpoints ────────────────────────────────────────────────

func apiUploadGetConfig(w http.ResponseWriter, r *http.Request) {
	key, ok := requireEventKey(r, w)
	if !ok {
		return
	}
	m, err := getOrCreateManager(key)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, m.store.snapshot().Config)
}

func apiUploadSaveConfig(w http.ResponseWriter, r *http.Request) {
	key, ok := requireEventKey(r, w)
	if !ok {
		return
	}
	var cfg eventConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	cfg.EventKey = key
	if cfg.TitleTemplate == "" {
		cfg.TitleTemplate = defaultTitleTemplate
	}
	if cfg.DescriptionTemplate == "" {
		cfg.DescriptionTemplate = defaultDescriptionTemplate
	}
	m, err := getOrCreateManager(key)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := m.store.update(func(s *eventState) {
		s.Config = cfg
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// Switching video_dir should trigger an immediate scan so the UI updates.
	go m.scanNow()
	writeJSON(w, cfg)
}

func apiUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	key, ok := requireEventKey(r, w)
	if !ok {
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeJSONError(w, http.StatusBadRequest, "parse multipart: "+err.Error())
		return
	}
	f, header, err := r.FormFile("file")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "missing 'file' field")
		return
	}
	defer f.Close()
	dir := filepath.Join(dataRoot(), "events", key)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".png"
	}
	out := filepath.Join(dir, "thumbnail"+ext)
	fout, err := os.Create(out)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer fout.Close()
	if _, err := io.Copy(fout, f); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	m, err := getOrCreateManager(key)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_ = m.store.update(func(s *eventState) {
		s.Config.ThumbnailPath = out
	})
	writeJSON(w, map[string]string{"thumbnail_path": out})
}

func apiUploadGetState(w http.ResponseWriter, r *http.Request) {
	key, ok := requireEventKey(r, w)
	if !ok {
		return
	}
	m, err := getOrCreateManager(key)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, m.store.snapshot())
}

func apiUploadScan(w http.ResponseWriter, r *http.Request) {
	key, ok := requireEventKey(r, w)
	if !ok {
		return
	}
	m, err := getOrCreateManager(key)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	m.scanNow()
	writeJSON(w, map[string]bool{"ok": true})
}

type filenameBody struct {
	Filename string `json:"filename"`
}

func apiUploadRetry(w http.ResponseWriter, r *http.Request) {
	key, ok := requireEventKey(r, w)
	if !ok {
		return
	}
	var body filenameBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Filename == "" {
		writeJSONError(w, http.StatusBadRequest, "filename required")
		return
	}
	m, err := getOrCreateManager(key)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := m.requestRetry(body.Filename); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	m.nudge()
	writeJSON(w, map[string]bool{"ok": true})
}

func apiUploadSkip(w http.ResponseWriter, r *http.Request) {
	key, ok := requireEventKey(r, w)
	if !ok {
		return
	}
	var body filenameBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Filename == "" {
		writeJSONError(w, http.StatusBadRequest, "filename required")
		return
	}
	m, err := getOrCreateManager(key)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := m.requestSkip(body.Filename); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]bool{"ok": true})
}

func apiUploadListProfiles(w http.ResponseWriter, r *http.Request) {
	profiles, err := listProfiles()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]any{"profiles": profiles})
}

func apiUploadProfileLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ProfileName string `json:"profile_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ProfileName == "" {
		writeJSONError(w, http.StatusBadRequest, "profile_name required")
		return
	}
	// Login blocks for as long as the browser stays open. Run it in a
	// goroutine so the HTTP request returns immediately; the operator
	// closes the window to finish. Use context.Background() because
	// r.Context() is canceled the moment we write the response.
	go func(name string) {
		if err := driver.Login(context.Background(), name); err != nil {
			log.Printf("login (%s): %v", name, err)
		}
	}(body.ProfileName)
	writeJSON(w, map[string]bool{"ok": true})
}

func apiUploadProfileCheck(w http.ResponseWriter, r *http.Request) {
	profile := r.URL.Query().Get("profile_name")
	if profile == "" {
		writeJSONError(w, http.StatusBadRequest, "profile_name required")
		return
	}
	name, err := driver.CheckChannel(r.Context(), profile)
	if err != nil {
		writeJSON(w, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, map[string]string{"channel_name": name})
}

// apiVideosManualSet records a manually-entered YT video ID for a TBA match.
func apiVideosManualSet(w http.ResponseWriter, r *http.Request) {
	key, ok := requireEventKey(r, w)
	if !ok {
		return
	}
	var body struct {
		MatchKey  string `json:"match_key"`
		VideoID   string `json:"yt_video_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.MatchKey == "" {
		writeJSONError(w, http.StatusBadRequest, "match_key required")
		return
	}
	m, err := getOrCreateManager(key)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := m.store.update(func(s *eventState) {
		if body.VideoID == "" {
			delete(s.ManualVideoIDs, body.MatchKey)
		} else {
			s.ManualVideoIDs[body.MatchKey] = body.VideoID
		}
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]bool{"ok": true})
}

func apiVideosManualDelete(w http.ResponseWriter, r *http.Request) {
	key, ok := requireEventKey(r, w)
	if !ok {
		return
	}
	matchKey := r.URL.Query().Get("match_key")
	if matchKey == "" {
		writeJSONError(w, http.StatusBadRequest, "match_key required")
		return
	}
	m, err := getOrCreateManager(key)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_ = m.store.update(func(s *eventState) {
		delete(s.ManualVideoIDs, matchKey)
	})
	writeJSON(w, map[string]bool{"ok": true})
}
