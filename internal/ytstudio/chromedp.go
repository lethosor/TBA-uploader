package ytstudio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// ChromedpDriver is the production Driver. Construct with NewChromedpDriver.
type ChromedpDriver struct {
	ProfileRoot string // directory containing one subdir per profile_name
	BrowserExe  string // optional explicit path; empty => autodetect
	Verbose     bool
}

// NewChromedpDriver returns a driver that stores profiles under profileRoot.
// If browserExe is "" the driver searches the usual install locations for
// Brave, then Edge, then Chrome.
func NewChromedpDriver(profileRoot, browserExe string) *ChromedpDriver {
	return &ChromedpDriver{ProfileRoot: profileRoot, BrowserExe: browserExe}
}

func (d *ChromedpDriver) logf(format string, args ...any) {
	if d.Verbose {
		log.Printf("ytstudio: "+format, args...)
	}
}

// findBrowser returns an absolute path to a Brave/Edge/Chrome binary.
func (d *ChromedpDriver) findBrowser() (string, error) {
	if d.BrowserExe != "" {
		return d.BrowserExe, nil
	}
	var candidates []string
	switch runtime.GOOS {
	case "windows":
		pf := os.Getenv("ProgramFiles")
		pf86 := os.Getenv("ProgramFiles(x86)")
		local := os.Getenv("LOCALAPPDATA")
		candidates = []string{
			filepath.Join(pf, "BraveSoftware", "Brave-Browser", "Application", "brave.exe"),
			filepath.Join(pf86, "BraveSoftware", "Brave-Browser", "Application", "brave.exe"),
			filepath.Join(local, "BraveSoftware", "Brave-Browser", "Application", "brave.exe"),
			filepath.Join(pf, "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(pf86, "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(pf, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(pf86, "Google", "Chrome", "Application", "chrome.exe"),
		}
	case "darwin":
		candidates = []string{
			"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		}
	default:
		candidates = []string{
			"/usr/bin/brave-browser",
			"/usr/bin/brave",
			"/opt/brave.com/brave/brave",
			"/snap/bin/brave",
			"/usr/bin/microsoft-edge",
			"/usr/bin/google-chrome",
			"/usr/bin/chromium",
		}
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", errors.New("ytstudio: no Brave/Edge/Chrome binary found (set browser_exe in config)")
}

// allocate returns a chromedp ExecAllocator bound to the given profile.
// Callers must defer the returned cancel func.
func (d *ChromedpDriver) allocate(ctx context.Context, profileName string, headless bool) (context.Context, context.CancelFunc, error) {
	if profileName == "" {
		return nil, nil, errors.New("ytstudio: profile_name is empty")
	}
	profileDir := filepath.Join(d.ProfileRoot, profileName)
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		return nil, nil, err
	}
	// Remove the SingletonLock left behind by an interrupted previous run.
	// chromedp's "cannot start, profile in use" failures all come back to
	// this file.
	_ = os.Remove(filepath.Join(profileDir, "SingletonLock"))

	exe, err := d.findBrowser()
	if err != nil {
		return nil, nil, err
	}

	opts := []chromedp.ExecAllocatorOption{
		chromedp.ExecPath(exe),
		chromedp.UserDataDir(profileDir),
		chromedp.NoSandbox,
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		// Kills the navigator.webdriver tell at the Blink level; the
		// applyStealth init script handles anything that flag misses.
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("no-restore-last-session", true),
		chromedp.Flag("restore-last-session", "false"),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.WindowSize(1400, 900),
	}
	if headless {
		// Headless Chrome advertises "HeadlessChrome/…" and YT Studio bounces
		// us to an unsupported-browser page. A spoofed UA fixes that path,
		// but the spoof itself is a fingerprintable mismatch against the
		// browser's Client Hints — so we only apply it here, not in headed
		// runs where Brave's native UA passes the check on its own.
		opts = append(opts,
			chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"),
			chromedp.Headless,
		)
	} else {
		opts = append(opts, chromedp.Flag("headless", false))
	}
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	return allocCtx, cancel, nil
}

// detectSignIn checks whether the current page URL indicates a sign-in
// redirect. Any URL on accounts.google.com or anything containing "/signin"
// signals an expired session.
func detectSignIn(currentURL string) bool {
	u := strings.ToLower(currentURL)
	return strings.Contains(u, "accounts.google") || strings.Contains(u, "/signin")
}

// Login opens YouTube Studio non-headless and blocks until the operator
// closes the browser window. The profile cookies persist after close.
func (d *ChromedpDriver) Login(ctx context.Context, profileName string) error {
	deadline := DefaultLoginDeadline
	ctx, cancelTO := context.WithTimeout(ctx, deadline)
	defer cancelTO()

	allocCtx, cancelAlloc, err := d.allocate(ctx, profileName, false)
	if err != nil {
		return err
	}
	defer cancelAlloc()

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()

	if err := chromedp.Run(browserCtx,
		applyStealth(),
		chromedp.Navigate("https://studio.youtube.com"),
	); err != nil {
		return err
	}
	d.logf("login: browser open, waiting for operator to close")
	// chromedp.NewContext registers a target-detached handler; when the
	// operator closes the window, browserCtx.Done() fires.
	<-browserCtx.Done()
	return nil
}

// CheckChannel opens YT Studio headlessly and reads the channel name. Returns
// ErrSessionExpired when the profile no longer has a session.
func (d *ChromedpDriver) CheckChannel(ctx context.Context, profileName string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	allocCtx, cancelAlloc, err := d.allocate(ctx, profileName, true)
	if err != nil {
		return "", err
	}
	defer cancelAlloc()
	bctx, cancelB := chromedp.NewContext(allocCtx)
	defer cancelB()

	var currentURL, channelName string
	err = chromedp.Run(bctx,
		applyStealth(),
		chromedp.Navigate("https://studio.youtube.com"),
		chromedp.Sleep(2*time.Second),
		chromedp.Location(&currentURL),
		chromedp.Evaluate(jsReadChannelName, &channelName),
	)
	if err != nil {
		return "", err
	}
	if detectSignIn(currentURL) {
		return "", ErrSessionExpired
	}
	channelName = strings.TrimSpace(channelName)
	return channelName, nil
}

// Upload performs the full upload: open Studio, click Create, attach the
// file via setInputFiles on the hidden <input type=file>, fill title and
// description, optionally set thumbnail, wait for "Checks complete", set
// visibility, extract the 11-char ID, Save, optionally add to playlist.
func (d *ChromedpDriver) Upload(ctx context.Context, profileName string, in UploadInput) (UploadResult, error) {
	ctx, cancelTO := context.WithTimeout(ctx, DefaultUploadDeadline)
	defer cancelTO()

	allocCtx, cancelAlloc, err := d.allocate(ctx, profileName, false)
	if err != nil {
		return UploadResult{}, err
	}
	defer cancelAlloc()
	bctx, cancelB := chromedp.NewContext(allocCtx)
	defer cancelB()

	visibility := strings.ToUpper(strings.TrimSpace(in.Visibility))
	if visibility == "" {
		visibility = "PUBLIC"
	}

	abs, err := filepath.Abs(in.VideoPath)
	if err != nil {
		return UploadResult{}, err
	}
	var thumbAbs string
	if in.ThumbnailPath != "" {
		thumbAbs, err = filepath.Abs(in.ThumbnailPath)
		if err != nil {
			return UploadResult{}, err
		}
	}

	var (
		currentURL  string
		channelName string
		videoID     string
	)

	// Step 1: navigate to YT Studio and check we're signed in.
	d.logf("step 1: navigate to studio")
	if err := chromedp.Run(bctx,
		applyStealth(),
		chromedp.Navigate("https://studio.youtube.com"),
		chromedp.Sleep(3*time.Second),
		chromedp.Location(&currentURL),
	); err != nil {
		return UploadResult{}, fmt.Errorf("open studio: %w", err)
	}
	d.logf("step 1: at %s", currentURL)
	if detectSignIn(currentURL) {
		return UploadResult{}, ErrSessionExpired
	}

	// Read channel name; non-fatal if missing.
	_ = chromedp.Run(bctx, chromedp.Evaluate(jsReadChannelName, &channelName))
	channelName = strings.TrimSpace(channelName)
	d.logf("step 1: channel=%q", channelName)

	// Step 2: open the Create dropdown, click "Upload videos", and feed the
	// file in via the CDP Page.fileChooserOpened interception API. The
	// dropdown's menu items live inside ytcp-text-menu's shadow root, so we
	// click them with a shadow-piercing JS walker. The file input itself is
	// also inside shadow DOM — interception bypasses both problems by
	// setting files via the backend node ID delivered with the chooser event.
	d.logf("step 2: arming file-chooser interceptor")
	if err := attachFileViaChooser(bctx, abs, d.logf); err != nil {
		return UploadResult{}, fmt.Errorf("attach file: %w", err)
	}
	d.logf("file attached: %s", abs)

	// Step 3: details — title, description.
	if err := chromedp.Run(bctx,
		chromedp.WaitVisible("#title-textarea", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		fillTextbox("#title-textarea #textbox", in.Title),
		fillTextbox("#description-textarea #textbox", in.Description),
	); err != nil {
		return UploadResult{}, fmt.Errorf("fill details: %w", err)
	}

	// Step 4: thumbnail (optional). The thumbnail tile has its own hidden
	// <input type=file accept=image/*>; finding it by accept attribute keeps
	// us off the main video input.
	if thumbAbs != "" {
		d.logf("step 4: attaching thumbnail")
		if err := attachThumbnailViaChooser(bctx, thumbAbs, d.logf); err != nil {
			d.logf("thumbnail upload failed: %v (continuing)", err)
		} else {
			d.logf("step 4: thumbnail attached")
		}
	}

	// Step 5: wait for "Checks complete". This is the slowest step; YT can
	// take many minutes on large files.
	checksCtx, cancelChecks := context.WithTimeout(bctx, DefaultChecksCompleteDeadline)
	defer cancelChecks()
	if err := chromedp.Run(checksCtx,
		waitForText(`(?i)checks complete`),
	); err != nil {
		return UploadResult{}, fmt.Errorf("wait checks complete: %w", err)
	}
	d.logf("checks complete")

	// Step 6: walk to the Visibility step. The dialog uses test-id buttons.
	if err := chromedp.Run(bctx,
		jsClick(`button[test-id='VIDEO_ELEMENTS']`),
		chromedp.Sleep(500*time.Millisecond),
		jsClick(`button[test-id='REVIEW']`),
		chromedp.Sleep(500*time.Millisecond),
		jsClick(`button[test-id='REVIEW']`), // some flows need it twice to advance
		chromedp.Sleep(500*time.Millisecond),
		chromedp.WaitVisible("tp-yt-paper-radio-button", chromedp.ByQuery),
		jsClick(fmt.Sprintf(`tp-yt-paper-radio-button[name='%s']`, visibility)),
		chromedp.Sleep(500*time.Millisecond),
	); err != nil {
		return UploadResult{}, fmt.Errorf("advance to visibility: %w", err)
	}

	// Step 7: capture the 11-char video ID from the dialog before saving.
	_ = chromedp.Run(bctx,
		chromedp.Evaluate(jsExtractVideoID, &videoID),
	)
	d.logf("captured video id (pre-save): %q", videoID)

	// Step 8: click Save.
	if err := chromedp.Run(bctx,
		jsClick(`ytcp-button[id='done-button'], button[aria-label='Save']:not([disabled])`),
		chromedp.Sleep(3*time.Second),
	); err != nil {
		return UploadResult{}, fmt.Errorf("save: %w", err)
	}

	// Fallback ID recovery if not captured pre-save.
	if videoID == "" {
		_ = chromedp.Run(bctx,
			chromedp.Sleep(2*time.Second),
			chromedp.Evaluate(jsExtractVideoIDFallback, &videoID),
		)
		d.logf("captured video id (fallback): %q", videoID)
	}

	if videoID == "" {
		return UploadResult{ChannelName: channelName}, errors.New("could not extract video id from YT Studio")
	}

	// Step 9: add to playlist (optional). Reopens the edit dialog.
	if in.PlaylistName != "" {
		if err := d.addToPlaylist(bctx, videoID, in.PlaylistName); err != nil {
			d.logf("add-to-playlist failed: %v", err)
			// Non-fatal — operator can fix manually.
		} else {
			d.logf("added to playlist %q", in.PlaylistName)
		}
	}

	return UploadResult{VideoID: videoID, ChannelName: channelName}, nil
}

// addToPlaylist opens https://studio.youtube.com/video/<id>/edit, opens the
// playlist dropdown, ticks the checkbox whose label exactly matches
// playlistName, then commits.
//
// YouTube Studio's DOM doesn't store playlist IDs anywhere reachable from
// JS (neither as DOM attributes nor as Polymer/Lit properties on the
// checkbox rows). Matching by visible name is the only path that works.
// All clicks use shadow-pierce locate + native MouseClickXY because
// ytcp-button / ytcp-dropdown-trigger / ytcp-checkbox-lit all ignore
// synthetic el.click() (Polymer gesture system listens for the full
// pointerdown/up sequence).
func (d *ChromedpDriver) addToPlaylist(ctx context.Context, videoID, playlistName string) error {
	editURL := fmt.Sprintf("https://studio.youtube.com/video/%s/edit", videoID)
	if err := chromedp.Run(ctx,
		chromedp.Navigate(editURL),
		chromedp.WaitVisible("#title-textarea", chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
	); err != nil {
		return fmt.Errorf("open edit page: %w", err)
	}

	// Step a: locate and coord-click the playlist dropdown trigger.
	d.logf("playlist: locate dropdown trigger")
	var tx, ty float64
	var trigInfo string
	if err := chromedp.Run(ctx, shadowLocateBySelector(
		[]string{"ytcp-dropdown-trigger[use-placeholder]", "ytcp-dropdown-trigger"},
		&tx, &ty, &trigInfo,
	)); err != nil {
		return fmt.Errorf("locate dropdown trigger: %w", err)
	}
	if trigInfo == "" {
		return errors.New("playlist dropdown trigger not found")
	}
	d.logf("playlist: trigger at (%.0f,%.0f) %s", tx, ty, trigInfo)
	if err := chromedp.Run(ctx,
		humanSleep(150*time.Millisecond, 400*time.Millisecond),
		humanClick(tx, ty),
		humanSleep(1300*time.Millisecond, 1800*time.Millisecond),
	); err != nil {
		return fmt.Errorf("click dropdown trigger: %w", err)
	}
	d.logf("playlist: visible-checkbox count=%s", probeDropdownState(ctx))

	// Step a.5: type the playlist name into the dropdown's search input.
	// The dialog is virtualized — on channels with many playlists, most rows
	// render as DOM stubs with no text until scrolled into view, so a direct
	// name match can't find them. Filtering collapses the list to the row
	// we want before the matcher runs.
	if searchInfo, err := playlistFilterByName(ctx, playlistName); err != nil {
		d.logf("playlist: filter input failed: %v (proceeding without filter)", err)
	} else if searchInfo == "" {
		d.logf("playlist: no search input found, proceeding without filter")
	} else {
		d.logf("playlist: filtered via %s", searchInfo)
		if err := chromedp.Run(ctx, humanSleep(1200*time.Millisecond, 1800*time.Millisecond)); err != nil {
			return err
		}
		d.logf("playlist: post-filter count=%s", probeDropdownState(ctx))
	}

	// Step b: find the checkbox row whose label exactly matches playlistName
	// and coord-click it. We click the LABEL element wrapping the checkbox,
	// not the checkbox-lit itself — clicking the label is what the user does
	// and native HTML label-for-checkbox semantics make it the most reliable
	// way to toggle a Polymer ytcp-checkbox-lit.
	d.logf("playlist: locate row %q", playlistName)
	var cx, cy float64
	var rowInfo string
	if err := chromedp.Run(ctx, shadowLocatePlaylistRow(playlistName, &cx, &cy, &rowInfo)); err != nil {
		return fmt.Errorf("locate playlist row: %w", err)
	}
	if rowInfo == "" {
		return fmt.Errorf("playlist %q not found in dropdown", playlistName)
	}
	d.logf("playlist: row at (%.0f,%.0f) %s", cx, cy, rowInfo)
	if err := chromedp.Run(ctx,
		humanSleep(200*time.Millisecond, 500*time.Millisecond),
		humanClick(cx, cy),
		humanSleep(700*time.Millisecond, 1100*time.Millisecond),
	); err != nil {
		return fmt.Errorf("click playlist row: %w", err)
	}
	d.logf("playlist: post-row state=%s", probeRowChecked(ctx, playlistName))

	// Step c: click Done to close the dropdown.
	d.logf("playlist: locate Done")
	var dx, dy float64
	var doneInfo string
	if err := chromedp.Run(ctx, shadowLocateByText(
		[]string{"ytcp-button[test-id='done-button']", "ytcp-button", "button"},
		regexp.MustCompile(`(?i)^done$`),
		&dx, &dy, &doneInfo,
	)); err != nil {
		return fmt.Errorf("locate done: %w", err)
	}
	if doneInfo == "" {
		return errors.New("Done button not found")
	}
	d.logf("playlist: Done at (%.0f,%.0f) %s", dx, dy, doneInfo)
	if err := chromedp.Run(ctx,
		humanSleep(200*time.Millisecond, 500*time.Millisecond),
		humanClick(dx, dy),
	); err != nil {
		return fmt.Errorf("click done: %w", err)
	}
	// The playlist dialog is a tp-yt-paper-dialog with an animated close.
	// If we click Save while it's still fading out, the click hits the
	// backdrop instead. Poll until the dialog is gone.
	if err := waitDialogHidden(ctx, 5*time.Second); err != nil {
		d.logf("playlist: dialog-hidden wait: %v", err)
	} else {
		d.logf("playlist: dialog closed")
	}

	// Step d: click Save on the edit dialog to commit the change.
	d.logf("playlist: locate Save")
	var saveX, saveY float64
	var saveInfo string
	if err := chromedp.Run(ctx, shadowLocateBySelector(
		[]string{
			"ytcp-button#save",
			"ytcp-button[id='save']",
			"ytcp-button[test-id='SAVE']",
			"button[aria-label='Save']:not([aria-disabled='true'])",
		},
		&saveX, &saveY, &saveInfo,
	)); err != nil {
		return fmt.Errorf("locate save: %w", err)
	}
	if saveInfo == "" {
		return errors.New("Save button not found")
	}
	d.logf("playlist: Save at (%.0f,%.0f) %s", saveX, saveY, saveInfo)
	d.logf("playlist: pre-save trigger-text=%q save-enabled=%s", probeTriggerText(ctx), probeSaveEnabled(ctx))
	if err := chromedp.Run(ctx,
		humanSleep(300*time.Millisecond, 700*time.Millisecond),
		humanClick(saveX, saveY),
		chromedp.Sleep(4*time.Second),
	); err != nil {
		return fmt.Errorf("click save: %w", err)
	}
	d.logf("playlist: post-save trigger-text=%q save-enabled=%s", probeTriggerText(ctx), probeSaveEnabled(ctx))
	return nil
}

// probeTriggerText returns the visible text of the playlist dropdown trigger.
// After Done commits a playlist toggle, the trigger should show the playlist
// name (e.g. "Test") rather than the placeholder "Select".
func probeTriggerText(ctx context.Context) string {
	const js = `(() => {
		function walk(root, fn) {
			fn(root);
			const all = root.querySelectorAll('*');
			for (const el of all) {
				if (el.shadowRoot) walk(el.shadowRoot, fn);
			}
		}
		let t = null;
		walk(document, root => {
			if (t) return;
			const cands = [
				...root.querySelectorAll('ytcp-dropdown-trigger[use-placeholder]'),
				...root.querySelectorAll('ytcp-dropdown-trigger'),
			];
			const it = cands.find(e => e.offsetParent !== null);
			if (it) t = it;
		});
		if (!t) return '';
		return (t.innerText || t.textContent || '').trim().slice(0, 80);
	})()`
	var out string
	_ = chromedp.Run(ctx, chromedp.Evaluate(js, &out))
	return out
}

// waitDialogHidden polls until any ytcp-playlist-dialog is no longer visible
// (or its inner tp-yt-paper-dialog is closed). Returns nil on success, or an
// error on timeout.
func waitDialogHidden(ctx context.Context, timeout time.Duration) error {
	const js = `(() => {
		function walk(root, fn) {
			fn(root);
			const all = root.querySelectorAll('*');
			for (const el of all) {
				if (el.shadowRoot) walk(el.shadowRoot, fn);
			}
		}
		let visible = false;
		walk(document, root => {
			if (visible) return;
			root.querySelectorAll('ytcp-playlist-dialog tp-yt-paper-dialog, ytcp-playlist-dialog').forEach(el => {
				if (el.offsetParent !== null) visible = true;
			});
		});
		return visible;
	})()`
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var visible bool
		if err := chromedp.Run(ctx, chromedp.Evaluate(js, &visible)); err != nil {
			return err
		}
		if !visible {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(150 * time.Millisecond):
		}
	}
	return errors.New("playlist dialog still visible")
}

// probeDropdownState returns how many ytcp-checkbox-lit elements are visible
// anywhere on the page (piercing shadow). Used to confirm the playlist
// dropdown actually opened after the trigger click.
func probeDropdownState(ctx context.Context) string {
	const js = `(() => {
		function walk(root, fn) {
			fn(root);
			const all = root.querySelectorAll('*');
			for (const el of all) {
				if (el.shadowRoot) walk(el.shadowRoot, fn);
			}
		}
		let n = 0, checked = 0;
		walk(document, root => {
			root.querySelectorAll('ytcp-checkbox-lit').forEach(el => {
				if (el.offsetParent === null) return;
				n++;
				if (el.getAttribute('aria-checked') === 'true' || el.hasAttribute('checked')) checked++;
			});
		});
		return n + ' visible, ' + checked + ' checked';
	})()`
	var out string
	_ = chromedp.Run(ctx, chromedp.Evaluate(js, &out))
	return out
}

// probeRowChecked returns whether the checkbox row labeled with `name` is
// checked. Helps diagnose whether the row click actually toggled the
// underlying ytcp-checkbox-lit.
func probeRowChecked(ctx context.Context, name string) string {
	js := fmt.Sprintf(`(() => {
		const want = %q.toLowerCase();
		function walk(root, fn) {
			fn(root);
			const all = root.querySelectorAll('*');
			for (const el of all) {
				if (el.shadowRoot) walk(el.shadowRoot, fn);
			}
		}
		let found = null;
		walk(document, root => {
			root.querySelectorAll('ytcp-checkbox-lit').forEach(el => {
				if (found || el.offsetParent === null) return;
				let p = el, text = '';
				for (let i = 0; i < 4 && p; i++) {
					const t = (p.innerText || p.textContent || '').trim();
					if (t) { text = t; break; }
					p = p.parentElement;
				}
				if (text.toLowerCase() === want) found = el;
			});
		});
		if (!found) return 'not-found';
		const ariaChecked = found.getAttribute('aria-checked');
		const hasChecked = found.hasAttribute('checked');
		const innerChecked = found.shadowRoot ? !!found.shadowRoot.querySelector('[checked]') : null;
		return 'aria=' + ariaChecked + ' attr=' + hasChecked + ' inner=' + innerChecked;
	})()`, name)
	var out string
	_ = chromedp.Run(ctx, chromedp.Evaluate(js, &out))
	return out
}

// probeSaveEnabled inspects the Save button's enabled/disabled state.
func probeSaveEnabled(ctx context.Context) string {
	const js = `(() => {
		function walk(root, fn) {
			fn(root);
			const all = root.querySelectorAll('*');
			for (const el of all) {
				if (el.shadowRoot) walk(el.shadowRoot, fn);
			}
		}
		let btn = null;
		walk(document, root => {
			if (btn) return;
			const cands = [
				...root.querySelectorAll('ytcp-button#save'),
				...root.querySelectorAll('ytcp-button[id=save]'),
			];
			const it = cands.find(e => e.offsetParent !== null);
			if (it) btn = it;
		});
		if (!btn) return 'no-save-button';
		return 'aria-disabled=' + btn.getAttribute('aria-disabled') +
			' disabled-attr=' + btn.hasAttribute('disabled');
	})()`
	var out string
	_ = chromedp.Run(ctx, chromedp.Evaluate(js, &out))
	return out
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// applyStealth registers a Page init script that runs before any page script
// on every navigation. It patches the most-fingerprinted automation tells:
// navigator.webdriver (true under CDP), the missing window.chrome object,
// the empty navigator.plugins array, and a singleton languages list. Paired
// with --disable-blink-features=AutomationControlled, this covers the trio
// of checks every "is this a headless browser?" snippet runs first.
func applyStealth() chromedp.Action {
	const script = `
		try {
			Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
		} catch (e) {}
		if (!window.chrome) { window.chrome = { runtime: {} }; }
		try {
			Object.defineProperty(navigator, 'plugins', {
				get: () => [
					{ name: 'Chrome PDF Plugin', filename: 'internal-pdf-viewer' },
					{ name: 'Chrome PDF Viewer', filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai' },
					{ name: 'Native Client', filename: 'internal-nacl-plugin' },
				],
			});
		} catch (e) {}
		try {
			Object.defineProperty(navigator, 'languages', {
				get: () => ['en-US', 'en'],
			});
		} catch (e) {}
	`
	return chromedp.ActionFunc(func(ctx context.Context) error {
		_, err := page.AddScriptToEvaluateOnNewDocument(script).Do(ctx)
		return err
	})
}

// humanSleep blocks for a uniformly-random duration in [min, max). Used at
// action boundaries to break up the dead-on-fixed-interval pattern that a
// timing fingerprinter would otherwise see.
func humanSleep(min, max time.Duration) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		d := min
		if max > min {
			d += time.Duration(rand.Int64N(int64(max - min)))
		}
		select {
		case <-time.After(d):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
}

// humanClick is a drop-in replacement for chromedp.MouseClickXY that first
// dispatches a short series of mouseMoved events from a small random offset,
// then issues a separate pressed/released pair with a human-ish dwell
// between them. chromedp.MouseClickXY sends pressed+released back-to-back at
// the exact target with no prior cursor motion — a near-zero-cost tell for
// any behaviour analysis. The motion + dwell here costs ~200–500 ms per
// click but removes one of the most obvious automation signals.
func humanClick(x, y float64) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		startX := x + rand.Float64()*120 - 60
		startY := y + rand.Float64()*120 - 60
		steps := 3 + rand.IntN(3)
		for i := 1; i <= steps; i++ {
			t := float64(i) / float64(steps)
			mx := startX + (x-startX)*t
			my := startY + (y-startY)*t
			if err := input.DispatchMouseEvent(input.MouseMoved, mx, my).Do(ctx); err != nil {
				return err
			}
			select {
			case <-time.After(time.Duration(15+rand.IntN(35)) * time.Millisecond):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		if err := input.DispatchMouseEvent(input.MousePressed, x, y).
			WithButton(input.Left).WithClickCount(1).Do(ctx); err != nil {
			return err
		}
		select {
		case <-time.After(time.Duration(50+rand.IntN(80)) * time.Millisecond):
		case <-ctx.Done():
			return ctx.Err()
		}
		return input.DispatchMouseEvent(input.MouseReleased, x, y).
			WithButton(input.Left).WithClickCount(1).Do(ctx)
	})
}

// clickByText finds the first element matching tagSelector whose textContent
// matches re, and clicks it. Implemented via Evaluate because chromedp's
// built-in selectors don't support text regex.
//
// The Go regex is converted to a JS RegExp; we strip the `(?i)` inline-flag
// prefix (JS doesn't support it) and always pass the 'i' flag to RegExp.
func clickByText(tagSelector string, re *regexp.Regexp) chromedp.Action {
	var unused string
	return clickByTextResult(tagSelector, re, &unused)
}

// clickByTextResult is the same as clickByText, but writes a short description
// of the matched element (or "" if none) into *info so the caller can log it.
func clickByTextResult(tagSelector string, re *regexp.Regexp, info *string) chromedp.Action {
	pat := strings.TrimPrefix(re.String(), "(?i)")
	js := fmt.Sprintf(`
		(() => {
			const re = new RegExp(%q, 'i');
			const els = [...document.querySelectorAll(%q)];
			// Prefer an exact text match before falling back to a partial match.
			const visible = els.filter(e => e.offsetParent !== null && re.test((e.textContent || '').trim()));
			const exact = visible.find(e => re.test((e.textContent || '').trim().split('\n')[0]));
			const el = exact || visible[0];
			if (!el) return '';
			const id = el.id ? '#' + el.id : '';
			const cls = el.className ? '.' + ('' + el.className).split(' ').filter(Boolean).slice(0, 3).join('.') : '';
			const txt = ((el.textContent || '').trim().slice(0, 40)).replace(/\s+/g, ' ');
			el.click();
			return el.tagName.toLowerCase() + id + cls + ' :: ' + JSON.stringify(txt);
		})()
	`, pat, tagSelector)
	return chromedp.Evaluate(js, info)
}

// attachFileViaChooser drives the Create → "Upload videos" menu flow and
// feeds the file in via CDP file-chooser interception. The menu item lives
// in ytcp-text-menu's shadow root, so we click it with a shadow-piercing
// walker; the file input also lives in shadow DOM but interception bypasses
// that by setting files via the backend node ID delivered with the chooser
// event.
// attachThumbnailViaChooser locates the ytcp-thumbnail-uploader element (the
// big "Upload thumbnail" tile in the upload dialog's Details tab), clicks it
// with a real coordinate click, and feeds the image path through file-chooser
// interception. Same pattern as the main video upload: shadow-piercing locate
// + native click + CDP chooser intercept.
func attachThumbnailViaChooser(ctx context.Context, filePath string, logf func(string, ...any)) error {
	chooserCh := make(chan cdp.BackendNodeID, 1)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if e, ok := ev.(*page.EventFileChooserOpened); ok {
			select {
			case chooserCh <- e.BackendNodeID:
			default:
			}
		}
	})

	if err := chromedp.Run(ctx, page.SetInterceptFileChooserDialog(true)); err != nil {
		return fmt.Errorf("enable chooser intercept: %w", err)
	}
	defer func() {
		_ = chromedp.Run(ctx, page.SetInterceptFileChooserDialog(false))
	}()

	logf("thumbnail: locate uploader tile")
	var tx, ty float64
	var info string
	if err := chromedp.Run(ctx, shadowLocateBySelector(
		[]string{"ytcp-thumbnail-uploader", "ytcp-thumbnail-uploader button", "button[aria-label*='thumbnail' i]"},
		&tx, &ty, &info,
	)); err != nil {
		return fmt.Errorf("locate thumbnail uploader: %w", err)
	}
	if info == "" {
		return errors.New("thumbnail uploader element not found")
	}
	logf("thumbnail: tile at (%.0f,%.0f) %s", tx, ty, info)

	if err := chromedp.Run(ctx,
		humanSleep(200*time.Millisecond, 500*time.Millisecond),
		humanClick(tx, ty),
	); err != nil {
		return fmt.Errorf("click thumbnail tile: %w", err)
	}

	select {
	case bnid := <-chooserCh:
		logf("thumbnail: chooser opened on backend node %d", bnid)
		return chromedp.Run(ctx, dom.SetFileInputFiles([]string{filePath}).WithBackendNodeID(bnid))
	case <-time.After(15 * time.Second):
		return errors.New("thumbnail file chooser never opened")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func attachFileViaChooser(ctx context.Context, filePath string, logf func(string, ...any)) error {
	// Listener captures the backend node ID of the file chooser when it opens.
	// Buffered so a slow consumer doesn't deadlock the CDP event dispatcher.
	chooserCh := make(chan cdp.BackendNodeID, 1)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if e, ok := ev.(*page.EventFileChooserOpened); ok {
			select {
			case chooserCh <- e.BackendNodeID:
			default:
			}
		}
	})

	if err := chromedp.Run(ctx, page.SetInterceptFileChooserDialog(true)); err != nil {
		return fmt.Errorf("enable chooser intercept: %w", err)
	}
	defer func() {
		_ = chromedp.Run(ctx, page.SetInterceptFileChooserDialog(false))
	}()

	logf("step 2: click Create")
	var createInfo string
	if err := chromedp.Run(ctx,
		// The top-bar Create button. The class is stable on current Studio.
		// Fall back to text-based match if the class moves.
		jsClickFirstResult([]string{
			`ytcp-button.ytcpAppHeaderCreateIcon`,
			`#create-icon-button`,
		}, &createInfo),
		chromedp.Sleep(700*time.Millisecond),
	); err != nil {
		return fmt.Errorf("click create: %w", err)
	}
	logf("step 2: Create matched %s", createInfo)

	logf("step 2: shadow-pierce click Upload videos")
	var uploadInfo string
	if err := chromedp.Run(ctx,
		shadowClickByText(
			[]string{"tp-yt-paper-item", "ytcp-text-menu-item", "[role=menuitem]"},
			regexp.MustCompile(`(?i)^upload\s*videos?$`),
			&uploadInfo,
		),
	); err != nil {
		return fmt.Errorf("click upload-videos menu item: %w", err)
	}
	logf("step 2: Upload-videos matched %s", uploadInfo)

	// The upload modal opens but doesn't auto-trigger the native file chooser.
	// We have to click the "Select files" button inside it. The button is a
	// ytcp-button (Polymer custom element) whose click handler listens for
	// the full pointerdown/pointerup gesture sequence — a synthetic .click()
	// in JS is ignored. So we shadow-walk to find the button's screen
	// coordinates, then issue a real CDP mouse click there.
	logf("step 2: locate SELECT FILES button")
	var sx, sy float64
	var selectInfo string
	if err := chromedp.Run(ctx,
		chromedp.Sleep(1*time.Second),
		shadowLocateByText(
			[]string{"ytcp-button#select-files-button", "ytcp-button", "button", "[role=button]"},
			// pit-podcast: r"select file|choose file" — match either phrasing,
			// substring (not anchored), because the rendered text may wrap.
			regexp.MustCompile(`(?i)select\s*files?|choose\s*files?`),
			&sx, &sy, &selectInfo,
		),
	); err != nil {
		return fmt.Errorf("locate select-files: %w", err)
	}
	logf("step 2: SELECT FILES at (%.0f,%.0f) %s", sx, sy, selectInfo)
	if selectInfo == "" {
		return errors.New("could not find SELECT FILES button")
	}
	if err := chromedp.Run(ctx,
		humanSleep(250*time.Millisecond, 600*time.Millisecond),
		humanClick(sx, sy),
	); err != nil {
		return fmt.Errorf("native click select-files: %w", err)
	}

	select {
	case bnid := <-chooserCh:
		logf("step 2: chooser opened on backend node %d", bnid)
		return chromedp.Run(ctx, dom.SetFileInputFiles([]string{filePath}).WithBackendNodeID(bnid))
	case <-time.After(30 * time.Second):
		return errors.New("file chooser never opened after clicking SELECT FILES")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// jsClickFirstResult tries each selector in order and clicks the first
// visible match. Writes a short description of the match into *info.
func jsClickFirstResult(selectors []string, info *string) chromedp.Action {
	// Build a JS array literal of selectors.
	parts := make([]string, len(selectors))
	for i, s := range selectors {
		parts[i] = fmt.Sprintf("%q", s)
	}
	arr := "[" + strings.Join(parts, ",") + "]"
	js := fmt.Sprintf(`
		(() => {
			const sels = %s;
			for (const sel of sels) {
				const els = [...document.querySelectorAll(sel)];
				const el = els.find(e => e.offsetParent !== null);
				if (el) {
					el.click();
					const id = el.id ? '#' + el.id : '';
					return sel + ' :: ' + el.tagName.toLowerCase() + id;
				}
			}
			return '';
		})()
	`, arr)
	return chromedp.Evaluate(js, info)
}

// shadowLocateBySelector walks all shadow roots and returns the center point
// of the first visible element matching any of the given CSS selectors.
func shadowLocateBySelector(selectors []string, x, y *float64, info *string) chromedp.Action {
	parts := make([]string, len(selectors))
	for i, s := range selectors {
		parts[i] = fmt.Sprintf("%q", s)
	}
	selArr := "[" + strings.Join(parts, ",") + "]"
	js := fmt.Sprintf(`
		(() => {
			const sels = %s;
			let target = null;
			function walk(root) {
				if (target) return;
				for (const sel of sels) {
					try {
						const els = [...root.querySelectorAll(sel)];
						const it = els.find(e => e.offsetParent !== null);
						if (it) { target = it; return; }
					} catch (e) {}
				}
				const all = root.querySelectorAll('*');
				for (const el of all) {
					if (el.shadowRoot) walk(el.shadowRoot);
					if (target) return;
				}
			}
			walk(document);
			if (!target) return JSON.stringify({info: ''});
			const r = target.getBoundingClientRect();
			const id = target.id ? '#' + target.id : '';
			return JSON.stringify({
				x: r.left + r.width / 2,
				y: r.top + r.height / 2,
				info: target.tagName.toLowerCase() + id,
			});
		})()
	`, selArr)
	return chromedp.ActionFunc(func(ctx context.Context) error {
		var raw string
		if err := chromedp.Evaluate(js, &raw).Do(ctx); err != nil {
			return err
		}
		var out struct {
			X, Y float64
			Info string
		}
		if err := json.Unmarshal([]byte(raw), &out); err != nil {
			return fmt.Errorf("parse locate result %q: %w", raw, err)
		}
		*x, *y, *info = out.X, out.Y, out.Info
		return nil
	})
}

// shadowLocatePlaylistRow walks the document looking for the playlist
// checkbox row whose label matches name. Returns the row's center
// coordinates so the caller can issue a real MouseClickXY (the checkboxes
// ignore synthetic .click()).
//
// Matching strategy, in priority order:
//  1. First non-empty line of the row's text equals name (case-insensitive).
//     The playlist title renders on line 1; subsequent lines are metadata
//     like "12 videos" or "Private".
//  2. Any line in the row's text equals name (case-insensitive). Covers
//     layouts where the title isn't on line 1.
//
// On failure, samples are returned in JSON so the caller can log what
// candidate row texts were actually present.
func shadowLocatePlaylistRow(name string, x, y *float64, info *string) chromedp.Action {
	js := fmt.Sprintf(`
		(() => {
			const wantName = %q.toLowerCase();
			function walk(root, fn) {
				fn(root);
				const all = root.querySelectorAll('*');
				for (const el of all) {
					if (el.shadowRoot) walk(el.shadowRoot, fn);
				}
			}
			let target = null;
			const samples = [];
			walk(document, root => {
				if (target) return;
				root.querySelectorAll('ytcp-checkbox-lit').forEach(el => {
					if (target || el.offsetParent === null) return;
					// Climb to the nearest ancestor with non-empty text.
					let p = el;
					let text = '';
					for (let i = 0; i < 4 && p; i++) {
						const t = (p.innerText || p.textContent || '').trim();
						if (t) { text = t; break; }
						p = p.parentElement;
					}
					if (!text) return;
					const lines = text.split('\n').map(s => s.trim()).filter(Boolean);
					const first = (lines[0] || '').toLowerCase();
					if (first === wantName) { target = el; return; }
					if (lines.some(l => l.toLowerCase() === wantName)) { target = el; return; }
					if (samples.length < 8) samples.push(lines.slice(0, 2).join(' / '));
				});
			});
			if (!target) return JSON.stringify({info: '', samples});
			const r = target.getBoundingClientRect();
			return JSON.stringify({
				x: r.left + r.width / 2,
				y: r.top + r.height / 2,
				info: target.tagName.toLowerCase() + (target.id ? '#' + target.id : ''),
			});
		})()
	`, name)
	return chromedp.ActionFunc(func(ctx context.Context) error {
		var raw string
		if err := chromedp.Evaluate(js, &raw).Do(ctx); err != nil {
			return err
		}
		var out struct {
			X, Y    float64
			Info    string
			Samples []string
		}
		if err := json.Unmarshal([]byte(raw), &out); err != nil {
			return fmt.Errorf("parse locate result %q: %w", raw, err)
		}
		*x, *y = out.X, out.Y
		if out.Info == "" && len(out.Samples) > 0 {
			*info = ""
			// Pack samples into the trailing error path via a sentinel:
			// the caller logs *info verbatim on success, but on failure
			// it tests info == "". We surface samples through the locate
			// log too by appending them under a SAMPLES: prefix below.
			return fmt.Errorf("no row matched; sample rows seen: %v", out.Samples)
		}
		*info = out.Info
		return nil
	})
}

// playlistFilterByName locates the playlist dropdown's search input,
// click-focuses it via a real pointer event, then types `name` through the
// CDP Input.insertText command. The dialog's filter listens for genuine
// keyboard input events (not the synthetic setter-+-event pattern that works
// for React), so we have to drive the input from above the DOM. Returns ""
// if no visible search input is found, or a "tag :: placeholder" string for
// logging on success.
func playlistFilterByName(ctx context.Context, name string) (string, error) {
	const js = `
		(() => {
			function walk(root, fn) {
				fn(root);
				const all = root.querySelectorAll('*');
				for (const el of all) {
					if (el.shadowRoot) walk(el.shadowRoot, fn);
				}
			}
			let input = null;
			walk(document, root => {
				if (input) return;
				const cands = [
					...root.querySelectorAll('ytcp-playlist-dialog input'),
					...root.querySelectorAll('input[placeholder*="search" i]'),
					...root.querySelectorAll('input[aria-label*="search" i]'),
				];
				const it = cands.find(el => el.offsetParent !== null);
				if (it) input = it;
			});
			if (!input) return JSON.stringify({info: ''});
			const r = input.getBoundingClientRect();
			return JSON.stringify({
				x: r.left + r.width / 2,
				y: r.top + r.height / 2,
				info: input.tagName.toLowerCase() +
					' :: ' + (input.placeholder || input.getAttribute('aria-label') || ''),
			});
		})()
	`
	var raw string
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &raw)); err != nil {
		return "", err
	}
	var out struct {
		X, Y float64
		Info string
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return "", fmt.Errorf("parse filter locate %q: %w", raw, err)
	}
	if out.Info == "" {
		return "", nil
	}
	if err := chromedp.Run(ctx,
		humanClick(out.X, out.Y),
		humanSleep(150*time.Millisecond, 300*time.Millisecond),
		input.InsertText(name),
	); err != nil {
		return out.Info, fmt.Errorf("insert text: %w", err)
	}
	return out.Info, nil
}

// shadowLocateByText walks all shadow roots to find an element matching any
// tagSelector whose textContent matches re, and writes the center of its
// bounding rect into *x, *y plus a short description into *info. If no
// match, *info is left empty so the caller can decide what to do.
func shadowLocateByText(tagSelectors []string, re *regexp.Regexp, x, y *float64, info *string) chromedp.Action {
	pat := strings.TrimPrefix(re.String(), "(?i)")
	parts := make([]string, len(tagSelectors))
	for i, s := range tagSelectors {
		parts[i] = fmt.Sprintf("%q", s)
	}
	selArr := "[" + strings.Join(parts, ",") + "]"
	js := fmt.Sprintf(`
		(() => {
			const re = new RegExp(%q, 'i');
			const sels = %s;
			let target = null;
			function walk(root) {
				if (target) return;
				for (const sel of sels) {
					const els = [...root.querySelectorAll(sel)];
					const it = els.find(e =>
						e.offsetParent !== null &&
						re.test((e.textContent || '').trim())
					);
					if (it) { target = it; return; }
				}
				const all = root.querySelectorAll('*');
				for (const el of all) {
					if (el.shadowRoot) walk(el.shadowRoot);
					if (target) return;
				}
			}
			walk(document);
			if (!target) return JSON.stringify({info: ''});
			const r = target.getBoundingClientRect();
			const id = target.id ? '#' + target.id : '';
			const txt = ((target.textContent || '').trim().slice(0, 40)).replace(/\s+/g, ' ');
			return JSON.stringify({
				x: r.left + r.width / 2,
				y: r.top + r.height / 2,
				info: target.tagName.toLowerCase() + id + ' :: ' + JSON.stringify(txt),
			});
		})()
	`, pat, selArr)
	return chromedp.ActionFunc(func(ctx context.Context) error {
		var raw string
		if err := chromedp.Evaluate(js, &raw).Do(ctx); err != nil {
			return err
		}
		var out struct {
			X, Y float64
			Info string
		}
		if err := json.Unmarshal([]byte(raw), &out); err != nil {
			return fmt.Errorf("parse locate result %q: %w", raw, err)
		}
		*x, *y, *info = out.X, out.Y, out.Info
		return nil
	})
}

// shadowClickByText walks all shadow roots looking for an element matching
// any tagSelector with textContent that matches re, and clicks it.
func shadowClickByText(tagSelectors []string, re *regexp.Regexp, info *string) chromedp.Action {
	pat := strings.TrimPrefix(re.String(), "(?i)")
	parts := make([]string, len(tagSelectors))
	for i, s := range tagSelectors {
		parts[i] = fmt.Sprintf("%q", s)
	}
	selArr := "[" + strings.Join(parts, ",") + "]"
	js := fmt.Sprintf(`
		(() => {
			const re = new RegExp(%q, 'i');
			const sels = %s;
			let target = null;
			function walk(root) {
				if (target) return;
				for (const sel of sels) {
					const els = [...root.querySelectorAll(sel)];
					const it = els.find(e =>
						e.offsetParent !== null &&
						re.test((e.textContent || '').trim())
					);
					if (it) { target = it; return; }
				}
				const all = root.querySelectorAll('*');
				for (const el of all) {
					if (el.shadowRoot) walk(el.shadowRoot);
					if (target) return;
				}
			}
			walk(document);
			if (!target) return '';
			const id = target.id ? '#' + target.id : '';
			const txt = ((target.textContent || '').trim().slice(0, 40)).replace(/\s+/g, ' ');
			target.click();
			return target.tagName.toLowerCase() + id + ' :: ' + JSON.stringify(txt);
		})()
	`, pat, selArr)
	return chromedp.Evaluate(js, info)
}

// waitForVisibleInput polls until selector matches a visible element. The
// built-in WaitVisible returns when the element is in the DOM, which fires
// too early for some Studio panels that mount detached.
func waitForVisibleInput(selector string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		deadline := time.Now().Add(60 * time.Second)
		for time.Now().Before(deadline) {
			var nodes []*cdp.Node
			err := chromedp.Nodes(selector, &nodes, chromedp.ByQueryAll, chromedp.AtLeast(0)).Do(ctx)
			if err == nil && len(nodes) > 0 {
				return nil
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(250 * time.Millisecond):
			}
		}
		return fmt.Errorf("timeout waiting for %s", selector)
	})
}

// fillTextbox clears a contenteditable element and types fresh text. YT
// Studio's title and description fields are contenteditable, not <textarea>,
// so chromedp.SendKeys-on-the-element with a Ctrl-A first works most reliably.
func fillTextbox(selector, value string) chromedp.Action {
	js := fmt.Sprintf(`
		(() => {
			const el = document.querySelector(%q);
			if (!el) return false;
			el.focus();
			document.execCommand('selectAll', false, null);
			document.execCommand('insertText', false, %q);
			return true;
		})()
	`, selector, value)
	var ok bool
	return chromedp.Tasks{
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.Evaluate(js, &ok),
		chromedp.Sleep(200 * time.Millisecond),
	}
}

// waitForText polls document.body.innerText against a regex. Accepts patterns
// in Go syntax; the `(?i)` inline flag is stripped because JS RegExp doesn't
// understand it (the 'i' flag is always set anyway).
func waitForText(pattern string) chromedp.Action {
	pattern = strings.TrimPrefix(pattern, "(?i)")
	return chromedp.ActionFunc(func(ctx context.Context) error {
		for {
			var found bool
			js := fmt.Sprintf(`new RegExp(%q, 'i').test(document.body.innerText)`, pattern)
			if err := chromedp.Evaluate(js, &found).Do(ctx); err != nil {
				return err
			}
			if found {
				return nil
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(3 * time.Second):
			}
		}
	})
}

// jsClick clicks the first element matching the CSS selector via JS, which
// bypasses pointer-event-blocking overlays that real-click intercepts.
func jsClick(selector string) chromedp.Action {
	js := fmt.Sprintf(`
		(() => {
			const el = document.querySelector(%q);
			if (el) { el.click(); return true; }
			return false;
		})()
	`, selector)
	var ok bool
	return chromedp.Evaluate(js, &ok)
}

// ─── JS snippets ──────────────────────────────────────────────────────────────

// jsReadChannelName returns the channel name visible in Studio's side nav, or
// "" if not signed in. The channel name lives inside a shadow DOM that
// querySelectorAll can't pierce, so we parse it out of document.body.innerText
// (which does pierce shadow boundaries). Sign-in is detected via the
// /channel/UC… URL prefix Studio uses after auth.
const jsReadChannelName = `
	(() => {
		if (!location.pathname.match(/\/channel\/UC[\w-]+/)) return '';
		const text = document.body.innerText || '';
		const m = text.match(/Your channel\s*\n\s*([^\n]+)/);
		return m ? m[1].trim() : '';
	})()
`

// jsExtractVideoID looks at the visible upload dialog for the YT URL link.
const jsExtractVideoID = `
	(() => {
		const links = [...document.querySelectorAll(
			"ytcp-video-info a[href*='youtu.be'], " +
			"ytcp-video-info a.video-url-fadeable, " +
			"a[href*='youtu.be/']"
		)];
		for (const l of links) {
			const m = (l.getAttribute('href') || '').match(/(?:youtu\.be\/|v=)([\w-]{11})/);
			if (m) return m[1];
		}
		return '';
	})()
`

// jsExtractVideoIDFallback runs after Save when the dialog is gone. The
// Studio "videos" page links each row to /video/<id>/edit.
const jsExtractVideoIDFallback = `
	(() => {
		const links = [...document.querySelectorAll("a[href*='studio.youtube.com/video/']")];
		for (const l of links) {
			const m = (l.getAttribute('href') || '').match(/\/video\/([\w-]{11})/);
			if (m) return m[1];
		}
		return '';
	})()
`
