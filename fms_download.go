package main

import (
    "errors"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "os"
    "path"
    "strings"
    "time"

    "github.com/PuerkitoBio/goquery"
)

var FMSConfig struct {
    FmsUrl string `json:"fms_url"`
    DataFolder string `json:"data_folder"`
    TbaUrl string `json:"tba_url"`
}

func checkFMSConnection() {
    // make sure the FMS server is running
    logger.Printf("Looking for FMS at %s...\n", FMSConfig.FmsUrl)
    client := http.Client{Timeout: 5 * time.Second}
    _, err := client.Get(FMSConfig.FmsUrl)
    if err != nil {
        logger.Println("Failed to connect to FMS!", err)
    } else {
        logger.Println("Found FMS")
    }
}

func downloadFile(folder string, filename string, url string, overwrite bool) (filepath string, ok bool, err error) {
    // return conditions:
    //      filepath: always
    //      ok: if the file exists now
    //      err: if the file was not downloaded
    err = os.MkdirAll(folder, os.ModePerm)
    if err != nil {
        logger.Printf("Failed to create folder: %s: %s\n", folder, err)
    }
    filepath = path.Join(folder, filename)
    ok = false
    if !overwrite {
        if fileExists(filepath) {
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
    url := fmt.Sprintf("%s/FieldMonitor/MatchesPartialByLevel?levelParam=%d", FMSConfig.FmsUrl, level)
    folder = path.Join(FMSConfig.DataFolder, folder, fmt.Sprintf("level%d", level))
    // ensure that the matches folder exists even if no matches are fetched
    os.MkdirAll(path.Join(folder, "matches"), os.ModePerm);

    filename, ok, err := downloadFile(folder, "match_list.html", url, true)
    if !ok {
        return nil, err
    }
    reader, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer reader.Close()
    dom, err := goquery.NewDocumentFromReader(reader)
    if err != nil {
        return nil, err
    }
    var files []string
    dom.Find("tbody tr").Each(func(i int, row *goquery.Selection) {
        match_url, ok := row.Find("a").First().Attr("href")
        if !ok {
            logger.Printf("Couldn't find link in row %d\n", i)
            return
        }
        match_url = FMSConfig.FmsUrl + match_url
        button := row.Find("button").First()
        button_text := strings.Replace(button.Text(), " ", "", -1)
        button_text = strings.Replace(button_text, "/", "-", -1)
        filename, ok, err := downloadFile(path.Join(folder, "matches"), button_text + ".html", match_url, !new_only)
        if !ok {
            logger.Printf("Failed to download %s: %s\n", button_text, err)
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

func downloadRankings() ([]byte, error) {
    request, err := http.NewRequest("GET", FMSConfig.FmsUrl + "/Pit/GetData", nil)
    if err != nil {
        return nil, err
    }
    request.Header.Add("Referer", FMSConfig.FmsUrl + "/Pit/Qual")
    client := http.Client{Timeout: 5 * time.Second}
    response, err := client.Do(request)
    if err != nil {
        return nil, err
    }
    return ioutil.ReadAll(response.Body)
}
