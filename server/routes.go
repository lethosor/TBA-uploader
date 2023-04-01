package main

import "net/http"

func apiRoot(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello world"))
}
