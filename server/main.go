package main

import (
    "flag"
    "fmt"
    "log"
    "net/http"
)

func main() {
    port := flag.Int("port", 8808, "web server port")
    flag.Parse()

    mux := http.NewServeMux()
    mux.HandleFunc("/api", apiRoot)

    log.Printf("Listening on port %d", *port)
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), mux))
}
