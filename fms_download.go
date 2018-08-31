package main

import (
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path"
    "strings"
    "time"

    "github.com/PuerkitoBio/goquery"
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

func downloadMatches(level int, folder string, new_only bool) error {
    url := fmt.Sprintf("%s/FieldMonitor/MatchesPartialByLevel?levelParam=%d", FMSServer, level)
    folder = path.Join(folder, fmt.Sprintf("level%d", level))
    filename, err := downloadFile(folder, "matches.html", url)
    if err != nil {
        return err
    }
    reader, err := os.Open(filename)
    if err != nil {
        return err
    }
    dom, err := goquery.NewDocumentFromReader(reader)
    if err != nil {
        return err
    }
    dom.Find("tr").Each(func(i int, row *goquery.Selection) {
        match_url, _ := row.Find("a").First().Attr("href")
        match_url = FMSServer + match_url
        button := row.Find("button").First()
        button_text := strings.Replace(button.Text(), " ", "", -1)
        button_text = strings.Replace(button_text, "/", "-", -1)
        downloadFile(path.Join(folder, "matches"), button_text + ".html", match_url)
    })
    return nil
}

func downloadNewMatches(level int, folder string) error {
    return downloadMatches(level, folder, true)
}

func downloadAllMatches(level int, folder string) error {
    return downloadMatches(level, folder, false)
}
