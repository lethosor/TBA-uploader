package main

import (
    "flag"
    "io"
    "log"
    "os"
    "path/filepath"
)

var Version = "dev"
var logger *log.Logger

func main() {
    exe, err := os.Executable()
    if err != nil {
        log.Fatalf("Could not find executable path: %s\n", err)
    }

    port := flag.Int("port", 8808, "web server port")
    fms_server := flag.String("fms-server", "http://10.0.100.5", "FMS server address (including protocol)")
    no_fms := flag.Bool("no-fms", false, "disable FMS connectivity")
    data_folder := flag.String("data-folder", filepath.Join(filepath.Dir(exe), "fms_data"), "FMS data destination folder")
    web_folder := flag.String("web-folder", "", "folder to serve files from (defaults to bundled files)")
    flag.Parse()

    FMSConfig.Server = *fms_server

    FMSConfig.DataFolder, err = filepath.Abs(*data_folder)
    if err != nil {
        log.Printf("WARNING: path normalization of \"%s\" failed: %s\n", *data_folder, err)
    }

    log_path := filepath.Join(FMSConfig.DataFolder, "tba-uploader.log")
    log_file, err := os.OpenFile(log_path, os.O_APPEND | os.O_CREATE | os.O_WRONLY, os.ModePerm)
    if err != nil {
        log.Printf("WARNING: cannot open log file \"%s\": %s\n", log_path, err)
    }
    log_file.Write([]byte("\n"))
    log_writer := io.MultiWriter(os.Stdout, log_file)
    logger = log.New(log_writer, "", log.Flags())

    logger.Printf("Version: %s\n", Version)
    logger.Printf("FMS data folder: %s\n", FMSConfig.DataFolder)
    logger.Printf("Logging to %s\n", log_path)

    web_folder_abs := ""
    if *web_folder != "" {
        web_folder_abs, err = filepath.Abs(*web_folder)
        if err != nil {
            logger.Printf("WARNING: path normalization of \"%s\" failed: %s", *web_folder, err)
        }
        logger.Printf("Serving HTML from %s\n", web_folder_abs)
    } else {
        logger.Printf("Serving bundled HTML\n")
    }

    os.Chdir(filepath.Dir(exe))
    cwd, _ := os.Getwd()
    logger.Printf("Running in %s\n", cwd)

    if !*no_fms {
        go checkFMSConnection()
    }
    RunWebServer(*port, web_folder_abs);
}
