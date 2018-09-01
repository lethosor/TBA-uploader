package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "path/filepath"
    "strconv"
    "strings"

    "github.com/gorilla/mux"
)

type AllianceSummary struct {
    Teams []int `json:"teams"`
    Score int `json:"score"`
    Rp int `json:"rp"`
}

type MatchSummary struct {
    MatchId string `json:"match_id"`
    Red AllianceSummary `json:"red"`
    Blue AllianceSummary `json:"blue"`
}

func getRequestEventParams(r *http.Request) (*eventParams, bool) {
    if len(r.Header.Get("X-Event")) > 0 && len(r.Header.Get("X-Auth")) > 0 && len(r.Header.Get("X-Secret")) > 0 {
        return &eventParams{
            event: r.Header.Get("X-Event"),
            auth: r.Header.Get("X-Auth"),
            secret: r.Header.Get("X-Secret"),
        }, true
    }
    return nil, false
}

func apiTBARequest(path string, w http.ResponseWriter, r *http.Request) {
    params, ok := getRequestEventParams(r)
    if !ok {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("missing event/auth API parameters"))
        return
    }
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(fmt.Sprintf("read failed: %s", err)))
        return
    }

    res, err := sendTBARequest(path, body, params)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(fmt.Sprintf("TBA request failed: %s", err)))
        return
    }

    if res.StatusCode != http.StatusOK {
        w.WriteHeader(http.StatusInternalServerError)
        res_body, _ := ioutil.ReadAll(res.Body)
        w.Write([]byte(fmt.Sprintf("TBA error %d: %s", res.StatusCode, res_body)))
        return
    }

    w.Write([]byte("ok"));
}

func jsVersion(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(fmt.Sprintf(";VERSION=\"%s\";", Version)))
}

func apiUploadAwards(w http.ResponseWriter, r *http.Request) {
    apiTBARequest("awards/update", w, r)
}

func apiUploadMatches(w http.ResponseWriter, r *http.Request) {
    apiTBARequest("matches/update", w, r)
}

func apiFetchMatches(w http.ResponseWriter, r *http.Request) {
    level, err := strconv.Atoi(r.URL.Query().Get("level"))
    download_all := (r.URL.Query().Get("all") != "")
    if err != nil || (level < 1 || level > 3) {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte(fmt.Sprintf("invalid level: %d", level)))
        return
    }
    var files []string
    if download_all {
        files, err = downloadAllMatches(level, "")
    } else {
        files, err = downloadNewMatches(level, "")
    }
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(fmt.Sprintf("match downloaded failed: %s", err)))
        return
    }

    info := make([]MatchSummary, 0)
    if files != nil {
        for i := 0; i < len(files); i++ {
            fname := filepath.Base(files[i])
            info = append(info, MatchSummary{
                MatchId: strings.TrimSuffix(fname, filepath.Ext(fname)),
                Red: AllianceSummary{
                    Teams: make([]int, 3),
                },
                Blue: AllianceSummary{
                    Teams: make([]int, 3),
                },
            })
        }
    }

    output, err := json.Marshal(info)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(fmt.Sprintf("json encode failed: %s", err)))
        return
    }

    w.Write(output)
}

func RunWebServer(port int, dev bool) {
    r := mux.NewRouter()
    var fs http.FileSystem
    if dev {
        fs = http.Dir("./web/")
    } else {
        fs = assetFS()
    }
    r.HandleFunc("/js/version.js", jsVersion)
    r.HandleFunc("/api/awards/upload", apiUploadAwards)
    r.HandleFunc("/api/matches/fetch", apiFetchMatches)
    r.HandleFunc("/api/matches/upload", apiUploadMatches)
    r.PathPrefix("/").Handler(http.FileServer(fs))
    addr := fmt.Sprintf(":%d", port)
    log.Printf("Serving on %s\n", addr);
    http.ListenAndServe(addr, r);
}
