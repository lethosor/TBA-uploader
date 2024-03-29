package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/lethosor/TBA-uploader/fms_parser"
)

func TestMain(m *testing.M) {
	FMSConfig.FmsUrl = "http://localhost:5555"
	FMSConfig.DataFolder = "fms_data_test"
	os.Exit(m.Run())
}

func skipCI(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Test server not implemented in CI")
	}
}

func TestDownloadMatches(t *testing.T) {
	skipCI(t)

	fmt.Println(getMatchDownloadPath(MATCH_LEVEL_QUAL, "all"))
	files, err := downloadNewMatches(MATCH_LEVEL_QUAL, "all")
	if err != nil {
		t.Error("downloadNewMatches: ", err)
	}
	fmt.Println("downloadNewMatches: Downloaded", len(files), "matches")
	for i := 0; i < 5 && i < len(files); i++ {
		fmt.Printf("files[%d] = %s\n", i, files[i])
	}

	files, err = downloadAllMatches(MATCH_LEVEL_QUAL, "all")
	if err != nil {
		t.Error("downloadAllMatches: ", err)
	}
	fmt.Println("downloadAllMatches: Downloaded", len(files), "matches")
	for i := 0; i < 5 && i < len(files); i++ {
		fmt.Printf("files[%d] = %s\n", i, files[i])
	}

	if len(files) == 0 {
		t.Fatal("no matches downloaded")
	}

	match_json, err := fms_parser.ParseHTMLtoJSON(2019, files[0], fms_parser.FMSParseConfig{Playoff: false})
	if err != nil {
		t.Error("ParseHTMLtoJSON: ", err)
	}
	out, _ := json.MarshalIndent(match_json, "", "  ")
	println(string(out))
}

func TestDownloadReports(t *testing.T) {
	skipCI(t)

	pages, err := downloadReport("ScheduleReportQualification")
	if err != nil {
		t.Error("downloadReport failed:", err)
	}

	if len(pages) != 2 {
		t.Error(fmt.Sprintf("wrong page count: got %d, expected %d", len(pages), 2))
	}
}
