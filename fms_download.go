package main

import (
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path"
    "time"

    // "github.com/PuerkitoBio/goquery"
)

var FMSServer string
var FMSDataFolder string

func checkFMSConnection() {
    // make sure the FMS server is running
    log.Printf("Looking for FMS at %s...\n", FMSServer)
    client := http.Client{Timeout: 5 * time.Second}
    _, err := client.Get(FMSServer)
    if err != nil {
        log.Println("Failed to connect to FMS!")
    } else {
        log.Println("Found FMS")
    }
}

func downloadFile(folder string, filename string, url string) (string, error) {
    folder = path.Join(FMSDataFolder, folder)
    os.MkdirAll(folder, os.ModePerm)
    filepath := path.Join(folder, filename)

    // Create the file
    out, err := os.Create(filepath)
    if err != nil {
        return filepath, err
    }
    defer out.Close()

    // Get the data
    resp, err := http.Get(url)
    if err != nil {
        return filepath, err
    }
    defer resp.Body.Close()

    // Write the body to file
    _, err = io.Copy(out, resp.Body)
    if err != nil {
        return filepath, err
    }

    return filepath, nil

}

func downloadMatches(level int, folder string, new_only bool) {
    url := fmt.Sprintf("%s/FieldMonitor/MatchesPartialByLevel?levelParam=%d", FMSServer, level)
    folder = path.Join(folder, fmt.Sprintf("level%d", level))
    downloadFile(folder, "matches.html", url)
}

func downloadNewMatches(level int, folder string) {
    downloadMatches(level, folder, true)
}

func downloadAllMatches(level int, folder string) {
    downloadMatches(level, folder, false)
}
