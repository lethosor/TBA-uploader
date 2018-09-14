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

func marshalFMSConfig(w http.ResponseWriter) ([]byte, error) {
    out, err := json.Marshal(FMSConfig)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        log.Printf("JSON.Marshal(FMSConfig): %s\n", err)
        return nil, err
    }
    return out, nil
}

func jsVersion(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(fmt.Sprintf("window.VERSION=\"%s\";", Version)))
}

func jsFMSConfig(w http.ResponseWriter, r *http.Request) {
    out, err := marshalFMSConfig(w)
    if err == nil {
        w.Write([]byte("window.FMS_CONFIG="))
        w.Write(out)
        w.Write([]byte(";"))
    }
}

func apiGetFMSConfig(w http.ResponseWriter, r *http.Request) {
    out, err := marshalFMSConfig(w)
    if err == nil {
        w.Write(out)
    }
}

func apiSetFMSConfig(w http.ResponseWriter, r *http.Request) {
    body, _ := ioutil.ReadAll(r.Body)
    err := json.Unmarshal(body, &FMSConfig)
    log.Printf("Changed FMS Config: Server = \"%s\", Data Folder = \"%s\"\n", FMSConfig.Server, FMSConfig.DataFolder)
    resp := make(map[string]interface{})
    resp["ok"] = (err == nil)
    if err != nil {
        resp["error"] = err.Error()
    }
    resp["config"] = FMSConfig
    out, err := json.Marshal(resp)
    w.Write(out)
    if err != nil {
        log.Printf("apiSetFMSConfig: Marshal failed: %s\n", err)
    }
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
    var match_folder = getMatchDownloadPath(level, r.URL.Query().Get("event"))
    var files []string
    if download_all {
        files, err = downloadAllMatches(level, r.URL.Query().Get("event"))
    } else {
        files, err = downloadNewMatches(level, r.URL.Query().Get("event"))
    }
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(fmt.Sprintf("match downloaded failed: %s", err)))
        return
    }

    if files != nil {
        for i := 0; i < len(files); i++ {
            log.Printf("Downloaded %s\n", files[i])
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

            match_info, err := ParseHTMLtoJSON(files[i], level == 3)
            if err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                w.Write([]byte(fmt.Sprintf("failed to parse %s: %s", fname, err)))
                return
            }

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

            // remove any receipts for newly-downloaded files
            fname_receipt := fname_trimmed + ".receipt"
            os.Remove(path.Join(folder, fname_receipt))
        }
    }

    match_json_list := make([]map[string]interface{}, 0)
    match_files, err := ioutil.ReadDir(match_folder)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(fmt.Sprintf("download folder %s scan failed: %s", match_folder, err)))
    }

    for _, json_file := range match_files {
        if (!strings.HasSuffix(json_file.Name(), ".json")) {
            continue
        }
        json_path := path.Join(match_folder, json_file.Name())
        receipt_path := strings.TrimSuffix(json_path, filepath.Ext(json_path)) + ".receipt"
        if _, err := os.Stat(receipt_path); err == nil {
            // receipt exists, match was already uploaded to TBA
            continue
        }

        match_info := make(map[string]interface{})
        contents, err := ioutil.ReadFile(json_path)
        err = json.Unmarshal(contents, &match_info)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            w.Write([]byte(fmt.Sprintf("failed to parse %s: %s", json_path, err)))
            return
        }

        match_info["_fms_id"] = strings.Split(json_file.Name(), ".")[0]
        match_json_list = append(match_json_list, match_info)
    }

    output, err := json.Marshal(match_json_list)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(fmt.Sprintf("json encode failed: %s", err)))
        return
    }

    w.Write(output)
}

func apiMarkMatchesUploaded(w http.ResponseWriter, r *http.Request) {
    params, ok := getRequestEventParams(r)
    if !ok {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("missing event/auth API parameters"))
        return
    }
    level, err := strconv.Atoi(r.URL.Query().Get("level"))
    if err != nil || (level < 1 || level > 3) {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte(fmt.Sprintf("invalid level: %d", level)))
        return
    }
    var match_folder = getMatchDownloadPath(level, params.event)
    match_ids := make([]string, 0)
    body, err := ioutil.ReadAll(r.Body)
    err = json.Unmarshal(body, &match_ids)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(fmt.Sprintf("failed to parse match ID list: %s", err)))
        return
    }
    for _, match_id := range match_ids {
        ioutil.WriteFile(path.Join(match_folder, match_id + ".receipt"), []byte(match_id), os.ModePerm)
    }
}

func RunWebServer(port int, web_folder string) {
    r := mux.NewRouter()
    var fs http.FileSystem
    if web_folder != "" {
        fs = http.Dir(web_folder)
    } else {
        fs = assetFS()
    }
    r.HandleFunc("/js/version.js", jsVersion)
    r.HandleFunc("/js/fms_config.js", jsFMSConfig)
    r.HandleFunc("/api/fms_config/get", apiGetFMSConfig)
    r.HandleFunc("/api/fms_config/set", apiSetFMSConfig)
    r.HandleFunc("/api/awards/upload", apiUploadAwards)
    r.HandleFunc("/api/matches/fetch", apiFetchMatches)
    r.HandleFunc("/api/matches/upload", apiUploadMatches)
    r.HandleFunc("/api/matches/mark_uploaded", apiMarkMatchesUploaded)
    r.PathPrefix("/").Handler(http.FileServer(fs))
    addr := fmt.Sprintf(":%d", port)
    log.Printf("Serving on %s\n", addr)
    err := http.ListenAndServe(addr, r)
    if err != nil {
        log.Fatalf("Could not start server: %s\n", err)
    }
}
