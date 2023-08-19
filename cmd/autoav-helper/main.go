package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

var settings struct {
	VideoDir string
}

type FileInfo struct {
	Name  string `json:"name"`
	Mtime int64  `json:"mtime"`
}

func main() {
	settings.VideoDir = "/tmp/videos"

	lock := sync.Mutex{}
	mux := http.NewServeMux()
	handle := func(method string, path string, handler func(w http.ResponseWriter, r *http.Request)) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			lock.Lock()
			defer lock.Unlock()
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Internal error: %v\n%s", err, debug.Stack())
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(fmt.Sprintf("Internal error: %v", err)))
				}
			}()

			log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
			w.Header().Set("access-control-allow-origin", "*")
			if method != r.Method {
				w.WriteHeader(http.StatusMethodNotAllowed)
				w.Write([]byte("method not allowed: " + r.Method))
				return
			}
			handler(w, r)
		})
	}
	handle(http.MethodGet, "/", handleRoot)
	handle(http.MethodPost, "/save", handleSaveSettings)
	handle(http.MethodGet, "/api/list", apiList)
	handle(http.MethodGet, "/api/rename", apiRename)

	addr := ":8807"
	log.Printf("listening on %s", addr)
	http.ListenAndServe(addr, mux)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html")
	w.Write([]byte(fmt.Sprintf(`
		<form action="/save" method="POST">
			<label>
				Video folder:
				<input name="VideoDir" value="%s">
			</label>
			<br><br>
			<input type="submit" value="Save">
		</form>
	`, html.EscapeString(settings.VideoDir))))
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

	all_files, err := ioutil.ReadDir(settings.VideoDir)
	if err != nil {
		panic(err)
	}

	files := make([]FileInfo, 0)
	for _, file := range all_files {
		if file.Mode().IsRegular() && strings.HasPrefix(file.Name(), prefix) && strings.HasSuffix(file.Name(), suffix) {
			files = append(files, FileInfo{
				Name:  file.Name(),
				Mtime: file.ModTime().Unix(),
			})
		}
	}

	out, err := json.Marshal(files)
	if err != nil {
		panic(err)
	}

	w.Write(out)
}

func fileExists(file_path string) bool {
	_, err := os.Stat(file_path)
	return !os.IsNotExist(err)
}

func apiRename(w http.ResponseWriter, r *http.Request) {
	old_name := r.URL.Query().Get("old_name")
	new_name := r.URL.Query().Get("new_name")

	if old_name == "" || new_name == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("old_name and new_name are required"))
		return
	}

	old_path := path.Join(settings.VideoDir, old_name)
	new_path := path.Join(settings.VideoDir, new_name)

	if !fileExists(old_path) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("old_path not found: " + old_path))
		return
	}

	if fileExists(new_path) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte("new_path already exists: " + new_path))
		return
	}

	for i := 1; i <= 5; i++ {
		if i > 1 {
			log.Printf("retrying %d/5...", i)
		}

		err := os.Rename(old_path, new_path)
		if err != nil {
			panic(err)
		}

		if fileExists(new_path) && !fileExists(old_path) {
			break
		}

		time.Sleep(1 * time.Second)
	}

	w.Write([]byte("{\"ok\": true}"))
}
