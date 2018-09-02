package main

import (
    "flag"
    "log"
    "os"
    "path/filepath"
)

var Version = "dev"

func main() {
    exe, err := os.Executable()
    if err != nil {
        log.Fatalf("Could not find executable path: %s\n", err)
    }

    log.Printf("Version: %s\n", Version)
    port := flag.Int("port", 8808, "web server port")
    fms_server := flag.String("fms-server", "http://10.0.100.5", "FMS server address (including protocol)")
    no_fms := flag.Bool("no-fms", false, "disable FMS connectivity")
    data_folder := flag.String("data-folder", filepath.Join(filepath.Dir(exe), "fms_data"), "FMS data destination folder")
    dev := flag.Bool("dev", false, "enable developer mode (serve files from disk)")
    flag.Parse()
    FMSServer = *fms_server
    FMSDataFolder, err := filepath.Abs(*data_folder)
    if err != nil {
        log.Printf("WARNING: path normalization failed: %s\n", err)
    }
    log.Printf("FMS data folder: %s\n", FMSDataFolder)

    os.Chdir(filepath.Dir(exe))
    cwd, _ := os.Getwd()
    log.Printf("Running in %s\n", cwd)

    if !*no_fms {
        go checkFMSConnection()
    }
    RunWebServer(*port, *dev);
}
