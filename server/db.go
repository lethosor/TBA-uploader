package main

import (
    "encoding/json"
    "fmt"
    "io/fs"
    "io/ioutil"
    "os"
    "path"
    "path/filepath"
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

func dbNormalizeKey(key string) (string, error) {
    err := dbValidateKey(key)
    if err != nil {
        return "", err
    }
    return path.Clean(key), nil
}

// private - key should be normalized
func _dbGetEntryPath(key string) string {
    if db_base_path == "" {
        panic("DB base path not set")
    }
    return path.Join(db_base_path, path.Clean(key))
}

func dbReadEntry(key string) ([]byte, error) {
    key, err := dbNormalizeKey(key)
    if err != nil {
        return nil, err
    }

    entry_path := _dbGetEntryPath(key)
    if !isFile(entry_path) {
        return nil, fmt.Errorf("key not found: %s", key)
    }

    value, err := ioutil.ReadFile(entry_path)
    if err != nil {
        return nil, fmt.Errorf("read failed: %s: %w", key, err)
    }

    if filepath.Ext(key) == ".json" && !json.Valid(value) {
        return nil, fmt.Errorf("invalid JSON at key: %s", key)
    }

    return value, nil
}

func _dbListPrefixIfMatching(prefix string, match func(fs.DirEntry) bool) (results []string, err error) {
    if prefix != "" {
        prefix, err = dbNormalizeKey(prefix)
        if err != nil {
            return nil, err
        }
    }

    results = []string{}

    files, err := os.ReadDir(_dbGetEntryPath(prefix))
    if err != nil {
        return results, nil
    }

    for _, file := range files {
        if match(file) {
            results = append(results, path.Join(prefix, file.Name()))
        }
    }

    return results, nil
}

func dbListPrefix(prefix string) ([]string, error) {
    return _dbListPrefixIfMatching(prefix, func(file fs.DirEntry) bool {
        return file.Type().IsRegular()
    })
}

func dbListSubprefixes(prefix string) ([]string, error) {
    return _dbListPrefixIfMatching(prefix, func(file fs.DirEntry) bool {
        return file.Type().IsDir()
    })
}

func dbReadAllPrefix(prefix string) (results map[string][]byte, errors map[string]error, err error) {
    keys, err := dbListPrefix(prefix)
    if err != nil {
        return
    }

    results = make(map[string][]byte)
    errors = make(map[string]error)

    for _, key := range keys {
        results[key], errors[key] = dbReadEntry(key)
    }
    return
}

func dbWriteEntry(key string, value []byte) error {
    key, err := dbNormalizeKey(key)
    if err != nil {
        return err
    }

    if filepath.Ext(key) == ".json" && !json.Valid(value) {
        return fmt.Errorf("invalid JSON for key: %s", key)
    }

    entry_path := _dbGetEntryPath(key)

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
