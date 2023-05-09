package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "path"
    "regexp"
)

const (
    DEFAULT_DIR_PERMISSION  = 0775
    DEFAULT_FILE_PERMISSION = 0664
)

var db_base_path string
var INVALID_KEY_PATTERN = regexp.MustCompile(`[^A-Za-z0-9/_\.-]`)

func dbValidateKey(key string) error {
    if len(key) < 1 {
        return fmt.Errorf("key cannot be empty")
    }
    if key[len(key)-1] == '/' {
        return fmt.Errorf("key cannot end with '/': %s", key)
    }
    loc := INVALID_KEY_PATTERN.FindIndex([]byte(key))
    if loc != nil {
        return fmt.Errorf("key contains invalid character at index %d: %s", loc[0], key)
    }
    return nil
}

func dbGetEntryPath(key string) string {
    if db_base_path == "" {
        panic("DB base path not set")
    }

    return path.Join(db_base_path, path.Clean(key)) + ".json"
}

func dbReadEntry(key string) ([]byte, error) {
    err := dbValidateKey(key)
    if err != nil {
        return nil, err
    }

    entry_path := dbGetEntryPath(key)
    if !isFile(entry_path) {
        return nil, fmt.Errorf("key not found: %s", key)
    }

    value, err := ioutil.ReadFile(entry_path)
    if err != nil {
        return nil, fmt.Errorf("read failed: %s: %w", key, err)
    }

    if !json.Valid(value) {
        return nil, fmt.Errorf("invalid JSON at key: %s", key)
    }

    return value, nil
}

func dbWriteEntry(key string, value []byte) error {
    err := dbValidateKey(key)
    if err != nil {
        return err
    }

    if !json.Valid(value) {
        return fmt.Errorf("invalid JSON for key: %s", key)
    }

    entry_path := dbGetEntryPath(key)

    err = os.MkdirAll(path.Dir(entry_path), DEFAULT_DIR_PERMISSION)
    if err != nil {
        return fmt.Errorf("failed to create directory for key: %s: %w", key, err)
    }

    err = ioutil.WriteFile(entry_path, value, DEFAULT_FILE_PERMISSION)
    if err != nil {
        return fmt.Errorf("write failed: %s: %w", key, err)
    }

    return nil
}
