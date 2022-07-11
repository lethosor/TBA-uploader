package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
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
	FmsUrl     string `json:"fms_url"`
	DataFolder string `json:"data_folder"`
	TbaUrl     string `json:"tba_url"`
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

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return
	}

	return filepath, true, nil
}

func getLevelDataPath(folder string, level int, subfolder string) string {
	return path.Join(FMSConfig.DataFolder, folder, fmt.Sprintf("level%d", level), subfolder)
}

func getMatchDownloadPath(level int, folder string) string {
	return getLevelDataPath(folder, level, "matches")
}

func getRankingDownloadPath(level int, folder string) string {
	return getLevelDataPath(folder, level, "rankings")
}

func downloadMatches(level int, folder string, new_only bool) ([]string, error) {
	url := fmt.Sprintf("%s/FieldMonitor/MatchesPartialByLevel?levelParam=%d", FMSConfig.FmsUrl, level)
	folder = path.Join(FMSConfig.DataFolder, folder, fmt.Sprintf("level%d", level))
	// ensure that the matches folder exists even if no matches are fetched
	matches_dir := path.Join(folder, "matches")
	backups_dir := path.Join(folder, "backups")
	os.MkdirAll(matches_dir, os.ModePerm)
	os.MkdirAll(backups_dir, os.ModePerm)

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
		filename, ok, err := downloadFile(matches_dir, button_text+".html", match_url, !new_only)
		if !ok {
			logger.Printf("Failed to download %s: %s\n", button_text, err)
		} else if err == nil {
			files = append(files, filename)

			var err error
			file_content, err := ioutil.ReadFile(filename)
			if err != nil {
				logger.Printf("Failed to hash %s: %s\n", filename, err)
			} else {
				hash := md5.Sum(file_content)
				dest := path.Join(backups_dir, fmt.Sprintf("%s-%x.html", button_text, hash))
				if !fileExists(dest) {
					err = copyFile(filename, dest, false)
					if err != nil {
						logger.Printf("Failed to back up %s to %s: %s\n", filename, dest, err)
					}
				}
			}
		}
	})
	return files, nil
}

func copyFile(src, dest string, overwrite bool) error {
	if !overwrite && fileExists(dest) {
		return errors.New(fmt.Sprintf("destination already exists: %s", dest))
	}
	content, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dest, content, os.ModePerm)
}

func downloadNewMatches(level int, folder string) ([]string, error) {
	return downloadMatches(level, folder, true)
}

func downloadAllMatches(level int, folder string) ([]string, error) {
	return downloadMatches(level, folder, false)
}

func downloadRankings(level int, folder string) ([]byte, error) {
	ranking_path := getRankingDownloadPath(level, folder)
	os.MkdirAll(ranking_path, os.ModePerm)
	request, err := http.NewRequest("GET", FMSConfig.FmsUrl+"/Pit/GetData", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Referer", FMSConfig.FmsUrl+"/Pit/Qual")
	client := http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	out, err := ioutil.ReadAll(response.Body)
	if err == nil {
		ranking_files, _ := listFilesWithExtension(ranking_path, "json")
		i := len(ranking_files) + 1
		var dest_filename string
		for {
			dest_filename = path.Join(ranking_path, fmt.Sprintf("%05d.json", i))
			if !fileExists(dest_filename) {
				break
			}
			i += 1
		}
		err = ioutil.WriteFile(dest_filename, out, os.ModePerm)
	}
	return out, err
}

type reportPage map[string]interface{}

func makeReportRequest(report_action, report_type string, headers map[string]string, body_fields map[string]interface{}) (map[string]interface{}, error) {
	custom_data_raw, _ := json.Marshal([]map[string]string{{
		"reportType": report_type,
	}})
	request_body := map[string]interface{}{
		"reportAction":    report_action,
		"controlId":       "sfreportviewer",
		"reportPath":      "",
		"reportServerUrl": "",
		"processingMode":  "local",
		"locale":          "en-US",
		"CustomData":      string(custom_data_raw),
	}
	if body_fields != nil {
		for field, value := range body_fields {
			request_body[field] = value
		}
	}
	body_encoded, err := json.Marshal(request_body)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", FMSConfig.FmsUrl+"/Reports/PostReportAction", bytes.NewReader(body_encoded))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	request.Header.Set("Accept-Language", "en-US,en;q=0.9")
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("DNT", "1")
	request.Header.Set("Origin", "http://10.0.100.5")
	request.Header.Set("Referer", "http://10.0.100.5/Reports/"+report_type)
	request.Header.Set("User-Agent", "Mozilla/5.0")
	if headers != nil {
		for header_name, header_value := range headers {
			request.Header.Set(header_name, header_value)
		}
	}

	client := http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	response_raw, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	out := make(map[string]interface{})
	err = json.Unmarshal(response_raw, &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func getReportToken(report_type string) (string, error) {
	response, err := makeReportRequest("ReportLoad", report_type, nil, nil)
	if err != nil {
		return "", errors.New(fmt.Sprintf("makeReportRequest: %s", err))
	}

	return readFromStringGenericMap[string](response, "reportViewerID")
}

func downloadReportPage(report_type string, page int, token string) (reportPage, error) {
	return makeReportRequest("GetPageModel", report_type, nil, map[string]interface{}{
		"dataRefresh":          true,
		"dataSources":          nil,
		"isPrint":              true,
		"pageindex":            page,
		"pageInit":             true,
		"parameters":           nil,
		"refresh":              false,
		"reportViewerClientId": token,
		"reportViewerToken":    token,
	})

}

func downloadReport(report_type string) ([]reportPage, error) {
	token, err := getReportToken(report_type)
	if err != nil {
		return nil, err
	}

	page1, err := downloadReportPage(report_type, 1, token)
	if err != nil {
		return nil, err
	}

	page_count, err := readFromStringGenericMap[float64](page1, "reportPageModel", "TotalPages")
	if err != nil {
		return nil, err
	}

	pages := []reportPage{page1}
	for i := 2; i <= int(page_count); i++ {
		next_page, err := downloadReportPage(report_type, i, token)
		if err != nil {
			return nil, err
		}
		pages = append(pages, next_page)
	}
	return pages, nil
}
