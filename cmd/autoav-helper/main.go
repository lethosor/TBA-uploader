package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
)

var settings struct {
	VideoDir string
}

type FileInfo struct {
	Name  string
	Mtime int64
}

func main() {
	settings.VideoDir = "/tmp/videos"

	mux := http.NewServeMux()
	handle := func(method string, path string, handler func(w http.ResponseWriter, r *http.Request)) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Internal error: %v\n%s", err, debug.Stack())
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(fmt.Sprintf("Internal error: %v", err)))
				}
			}()

			log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
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
