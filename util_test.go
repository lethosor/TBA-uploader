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

func TestReadFromStringGenericMap(t *testing.T) {
    m := map[string]interface{} {
        "a": "foo",
        "b": 2,
    }

    a, err := readFromStringGenericMap[string](m, "a")
    if a != "foo" {
        t.Error("string lookup failed:", err)
    }

    b, err := readFromStringGenericMap[int](m, "b")
    if b != 2 {
        t.Error("int lookup failed:", err)
    }

    c, err := readFromStringGenericMap[int](m, "c")
    if err == nil {
        t.Error("expected failed read, got:", c)
    }

    b2, err := readFromStringGenericMap[string](m, "b")
    if err == nil {
        t.Error("expected type mismatch, got", b2)
    }
}

func TestReadFromStringGenericMapRecursive(t *testing.T) {
    m := map[string]interface{} {
        "a": map[string]interface{} {
            "b": 2,
        },
    }

    b, err := readFromStringGenericMap[int](m, "a", "b")
    if b != 2 {
        t.Error("recursive int lookup failed:", err)
    }

    c, err := readFromStringGenericMap[int](m, "a", "b", "c")
    if err == nil {
        t.Error("expected failed read under int, got:", c)
    }
}
