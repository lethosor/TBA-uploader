package main

import (
    "flag"
    "log"
)

var Version = "dev"

func main() {
    log.Printf("Version: %s\n", Version)
    port := flag.Int("port", 8808, "web server port")
    fms_server := flag.String("fms-server", "http://10.0.100.5", "FMS server address (including protocol)")
    no_fms := flag.Bool("no-fms", false, "disable FMS connectivity")
    data_folder := flag.String("data-folder", "fms_data", "FMS data destination folder")
    dev := flag.Bool("dev", false, "enable developer mode (serve files from disk)")
    flag.Parse()
    FMSServer = *fms_server
    FMSDataFolder = *data_folder
    if !*no_fms {
        go checkFMSConnection()
    }
    RunWebServer(*port, *dev);
}
