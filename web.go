package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "net/http"

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
    r.HandleFunc("/api/matches/upload", apiUploadMatches)
    r.PathPrefix("/").Handler(http.FileServer(fs))
    addr := fmt.Sprintf(":%d", port)
    log.Printf("Serving on %s\n", addr);
    http.ListenAndServe(addr, r);
}
