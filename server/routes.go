package main

import "net/http"

func apiRegisterHandlers(mux *http.ServeMux, prefix string) {
    mux.HandleFunc(prefix, apiRoot)
}

func apiRoot(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello world"))
}
