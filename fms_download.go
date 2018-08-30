package main

import (
    "log"
    "net/http"
    "time"
)

func checkFMSConnection(fms_server string) {
    // make sure the FMS server is running
    log.Printf("Looking for FMS at %s...\n", fms_server)
    client := http.Client{Timeout: 5 * time.Second}
    _, err := client.Get(fms_server)
    if err != nil {
        log.Println("Failed to connect to FMS!")
    } else {
        log.Println("Found FMS")
    }
}
