package main

import (
    "errors"
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

var FMSConfig struct {
    Server string `json:"server"`
    DataFolder string `json:"data_folder"`
}

func checkFMSConnection() {
    // make sure the FMS server is running
    log.Printf("Looking for FMS at %s...\n", FMSConfig.Server)
    client := http.Client{Timeout: 5 * time.Second}
    _, err := client.Get(FMSConfig.Server)
    if err != nil {
        log.Println("Failed to connect to FMS!")
    } else {
        log.Println("Found FMS")
    }
}

func downloadFile(folder string, filename string, url string, overwrite bool) (filepath string, ok bool, err error) {
    // return conditions:
    //      filepath: always
    //      ok: if the file exists now
    //      err: if the file was not downloaded
    os.MkdirAll(folder, os.ModePerm)
    filepath = path.Join(folder, filename)
    ok = false
    if !overwrite {
        if _, err := os.Stat(filepath); err == nil {
            // exists, don't overwrite
            return filepath, true, errors.New("already downloaded")
        }
    }

    // Create the file
    out, err := os.Create(filepath)
    if err != nil {
        return
    }
    defer out.Close()

    // Get the data
    resp, err := http.Get(url)
    if err != nil {
        return
    }
    defer resp.Body.Close()

    // Write the body to file
    _, err = io.Copy(out, resp.Body)
    if err != nil {
        return
    }

    return filepath, true, nil
}

func getMatchDownloadPath(level int, folder string) string {
    return path.Join(FMSConfig.DataFolder, folder, fmt.Sprintf("level%d", level), "matches")
}

func downloadMatches(level int, folder string, new_only bool) ([]string, error) {
    url := fmt.Sprintf("%s/FieldMonitor/MatchesPartialByLevel?levelParam=%d", FMSConfig.Server, level)
    folder = path.Join(FMSConfig.DataFolder, folder, fmt.Sprintf("level%d", level))
    filename, ok, err := downloadFile(folder, "matches.html", url, true)
    if !ok {
        return nil, err
    }
    reader, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    dom, err := goquery.NewDocumentFromReader(reader)
    if err != nil {
        return nil, err
    }
    var files []string
    dom.Find("tbody tr").Each(func(i int, row *goquery.Selection) {
        match_url, ok := row.Find("a").First().Attr("href")
        if !ok {
            log.Printf("Couldn't find link in row %d\n", i)
            return
        }
        match_url = FMSConfig.Server + match_url
        button := row.Find("button").First()
        button_text := strings.Replace(button.Text(), " ", "", -1)
        button_text = strings.Replace(button_text, "/", "-", -1)
        filename, ok, err := downloadFile(path.Join(folder, "matches"), button_text + ".html", match_url, !new_only)
        if !ok {
            log.Printf("Failed to download %s: %s\n", button_text, err)
        } else if err == nil {
            files = append(files, filename)
        }
    })
    return files, nil
}

func downloadNewMatches(level int, folder string) ([]string, error) {
    return downloadMatches(level, folder, true)
}

func downloadAllMatches(level int, folder string) ([]string, error) {
    return downloadMatches(level, folder, false)
}
