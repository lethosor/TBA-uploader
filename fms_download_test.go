package main

import (
    "os"
    "testing"
)

func TestMain(m *testing.M) {
    FMSServer = "http://localhost:5000"
    FMSDataFolder = "fms_data_test"
    os.Exit(m.Run())
}

func TestDownloadAllMatches(t *testing.T) {
    err := downloadAllMatches(2, "all")
    if err != nil {
        t.Error("downloadAllMatches: ", err)
    }
}
