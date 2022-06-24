package main

import (
    "errors"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
)

// list all files in dir with the given extension (no ".")
func listFilesWithExtension(dirname, extension string) ([]os.FileInfo, error){
    all_files, err := ioutil.ReadDir(dirname)
    if err != nil {
        return nil, err
    }
    files := make([]os.FileInfo, 0)
    for _, file := range all_files {
        if file.Mode().IsRegular() && filepath.Ext(file.Name()) == "." + extension {
            files = append(files, file)
        }
    }
    return files, nil
}

// no "." in new_extension
func replaceExtension(filename, new_extension string) string {
    return strings.TrimSuffix(filename, filepath.Ext(filename)) + "." + new_extension
}

func fileExists(filename string) bool {
    _, err := os.Stat(filename)
    return err == nil
}

func isFile(filename string) bool {
    info, err := os.Stat(filename)
    return err == nil && info.Mode().IsRegular()
}

func isDir(filename string) bool {
    info, err := os.Stat(filename)
    return err == nil && info.Mode().IsDir()
}

func readFromStringGenericMap[T any](m map[string]interface{}, keys ...string) (result T, err error) {
    for i, key := range keys {
        if raw_val, ok := m[key]; ok {
            if i < len(keys) - 1 {
                if m, ok = raw_val.(map[string]interface{}); ok {
                    continue
                } else {
                    err = errors.New(fmt.Sprintf("not a nested map: %s", key))
                    return
                }
            } else {
                // last value
                if val, ok := raw_val.(T); ok {
                    return val, nil
                } else {
                    err = errors.New(fmt.Sprintf("cannot convert key from %T to %T: %s", raw_val, result, key))
                    return
                }
            }
        } else {
            err = errors.New(fmt.Sprintf("key not present: %s", key))
            return
        }
    }
    err = errors.New("readFromStringGenericMap: unexpected end")
    return
}
