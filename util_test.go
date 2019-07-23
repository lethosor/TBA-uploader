package main

import (
    "fmt"
    "testing"
)

func TestListFilesWithExtension(t *testing.T) {
    files, err := listFilesWithExtension(".", "go")
    if err != nil {
        t.Error("listFilesWithExtension: ", err)
    }
    for _, file := range files {
        fmt.Println(file.Name())
    }
}

func TestReplaceExtension(t *testing.T) {
    res := replaceExtension("foo.json", "html")
    fmt.Println(res)
    if res != "foo.html" {
        t.Error("replaceExtension(foo.json, html): ", res)
    }
}
