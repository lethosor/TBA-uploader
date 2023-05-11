package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
)

func apiRegisterHandlers(mux *http.ServeMux, prefix string) {
    addRoute := func(path string, handler func(w http.ResponseWriter, r *http.Request)) {
        mux.HandleFunc(prefix+path, func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if r := recover(); r != nil {
                    switch err := r.(type) {
                    case APIError:
                        w.WriteHeader(err.code)
                        w.Write([]byte(err.message))
                    default:
                        w.WriteHeader(http.StatusInternalServerError)
                        w.Write([]byte(fmt.Sprintf("Unknown internal error: %v", err)))
                    }
                }
            }()
            handler(w, r)
        })

    }

    addRoute("", apiRoot)
    addRoute("/db/read", apiDbRead)
    addRoute("/db/write", apiDbWrite)
}

type APIError struct {
    code    int
    message string
}

func apiPanicCode(code int, message string, args ...interface{}) {
    logger.Printf("API Error %d: "+message, append([]interface{}{code}, args...)...)
    panic(APIError{
        code:    code,
        message: fmt.Sprintf(message, args...),
    })
}

func apiPanicBadRequest(message string, args ...interface{}) {
    apiPanicCode(http.StatusBadRequest, message, args...)
}

func apiPanicInternal(message string, args ...interface{}) {
    apiPanicCode(http.StatusInternalServerError, message, args...)
}

func apiRequireMethod(r *http.Request, method string) {
    if r.Method != method {
        apiPanicBadRequest("invalid method %s, expected %s", r.Method, method)
    }
}

func apiRequireQueryParam(r *http.Request, param string) string {
    res := r.URL.Query().Get(param)
    if res == "" {
        apiPanicBadRequest("missing parameter: %s", param)
    }
    return res
}

func apiRequireBody(r *http.Request) []byte {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        apiPanicInternal("body read failed: %v", err)
    }
    return body
}

func apiRoot(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("hello world"))
}

func apiDbRead(w http.ResponseWriter, r *http.Request) {
    key := apiRequireQueryParam(r, "key")
    value, err := dbReadEntry(key)
    if err != nil {
        apiPanicBadRequest("%v", err)
    }
    w.Write(value)
}

func apiDbWrite(w http.ResponseWriter, r *http.Request) {
    apiRequireMethod(r, http.MethodPost)
    key := apiRequireQueryParam(r, "key")
    body := apiRequireBody(r)

    err := dbWriteEntry(key, body)
    if err != nil {
        apiPanicInternal("%v", err)
    }
    w.Write([]byte("ok"))
}
