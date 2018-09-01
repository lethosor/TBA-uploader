package main

import (
    "encoding/json"
    "fmt"
    "os"
    "testing"
)

func TestMain(m *testing.M) {
    FMSServer = "http://localhost:5000"
    FMSDataFolder = "fms_data_test"
    os.Exit(m.Run())
}

func TestDownloadMatches(t *testing.T) {
    files, err := downloadNewMatches(2, "all")
    if err != nil {
        t.Error("downloadNewMatches: ", err)
    }
    fmt.Println("downloadNewMatches: Downloaded", len(files), "matches")
    for i := 0; i < 5 && i < len(files); i++ {
        fmt.Printf("files[%d] = %s\n", i, files[i])
    }

    files, err = downloadAllMatches(2, "all")
    if err != nil {
        t.Error("downloadAllMatches: ", err)
    }
    fmt.Println("downloadAllMatches: Downloaded", len(files), "matches")
    for i := 0; i < 5 && i < len(files); i++ {
        fmt.Printf("files[%d] = %s\n", i, files[i])
    }

    if (len(files) == 0) {
        t.Fatal("no matches downloaded")
    }

    match_json, err := ParseHTMLtoJSON(files[0])
    if err != nil {
        t.Error("ParseHTMLtoJSON: ", err)
    }
    out, _ := json.MarshalIndent(match_json, "", "  ")
    println(string(out))
}
