package main

import (
    "flag"
)

func main() {
    port := flag.Int("port", 8808, "web server port")
    fms_server := flag.String("fms-server", "http://10.0.100.5", "FMS server address (including protocol)")
    no_fms := flag.Bool("no-fms", false, "disable FMS connectivity")
    flag.Parse()
    if !*no_fms {
        checkFMSConnection(*fms_server)
    }
    RunWebServer(*port);
}
