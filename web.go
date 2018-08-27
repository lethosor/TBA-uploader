package main

import (
    "fmt"
    "net/http"

    "github.com/gorilla/mux"
)

func RunWebServer(port int) {
    r := mux.NewRouter()
    fs := http.Dir("./web/")
    r.PathPrefix("/").Handler(http.FileServer(fs))
    addr := fmt.Sprintf(":%d", port)
    fmt.Printf("Serving on %s\n", addr);
    http.ListenAndServe(addr, r);
}
