package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "path"
    "path/filepath"
    "strconv"
    "strings"

    "github.com/gorilla/mux"
)

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

    info := make([]map[string]interface{}, 0)
    if files != nil {
        for i := 0; i < len(files); i++ {
            fname := filepath.Base(files[i])
            match_number, err := strconv.Atoi(strings.Split(fname, "-")[0])
            if err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                w.Write([]byte(fmt.Sprintf("%s: failed to parse match ID", fname)))
                return
            }
            folder := filepath.Dir(files[i])
            fname_trimmed := strings.TrimSuffix(fname, filepath.Ext(fname))
            fname_json := fname_trimmed + ".json"

            match_info, err := ParseHTMLtoJSON(files[i])
            if err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                w.Write([]byte(fmt.Sprintf("failed to parse %s: %s", fname, err)))
                return
            }
            score_bytes, err := json.Marshal(match_info["score_breakdown"])
            if err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                w.Write([]byte(fmt.Sprintf("%s: score_breakdown serialization failed: %s", fname, err)))
                return
            }
            match_info["score_breakdown"] = string(score_bytes)
            match_info["_fms_id"] = match_number

            if (level == 3) {
                // playoffs
                code := getTBAPlayoffCode(match_number)
                match_info["comp_level"] = code.level
                match_info["set_number"] = code.set
                match_info["match_number"] = code.match
            } else {
                match_info["comp_level"] = "qm"
                match_info["set_number"] = 1
                match_info["match_number"] = match_number
            }

            match_json, err := json.Marshal(match_info)
            if err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                w.Write([]byte(fmt.Sprintf("%s: JSON serialization failed %s", fname, err)))
                return
            }
            ioutil.WriteFile(path.Join(folder, fname_json), match_json, os.ModePerm)

            info = append(info, match_info)
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
