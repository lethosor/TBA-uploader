package main

import (
    "flag"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
)

var logger *log.Logger
var Version = "dev"

func logInit(log_path string) {
    log_file, err := os.OpenFile(log_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, DEFAULT_FILE_PERMISSION)
    if err != nil {
        log.Printf("WARNING: cannot open log file \"%s\": %s\n", log_path, err)
    }
    log_file.Write([]byte("\n"))
    log_writer := io.MultiWriter(os.Stdout, log_file)
    logger = log.New(log_writer, "", log.Flags())

    logger.Printf("*** TBA-uploader start ***\n")
    logger.Printf("Version: %s\n", Version)
    logger.Printf("Logging to %s\n", log_path)
}

func main() {
    self_exe_path, err := os.Executable()
    if err != nil {
        log.Fatalf("Could not find executable path: %s\n", err)
    }

    port := flag.Int("port", 8808, "web server port")
    data_folder := flag.String("data-folder", filepath.Join(filepath.Dir(self_exe_path), "fms_data"), "FMS data destination folder")
    flag.Parse()

    os.MkdirAll(*data_folder, DEFAULT_DIR_PERMISSION)
    db_base_path = *data_folder

    log_path := filepath.Join(filepath.Dir(self_exe_path), "tba-uploader.log")
    logInit(log_path)
    logger.Printf("Data folder: %s\n", *data_folder)

    mux := http.NewServeMux()
    apiRegisterHandlers(mux, "/api")

    log.Printf("Listening on port %d", *port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), mux))
}
