package main

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	db_base_path = "test/test_data"
	os.Exit(m.Run())
}

func clearDataDir(t *testing.T) {
	assert.Equal(t, "test/test_data", db_base_path) // safeguard before RemoveAll
	assert.NoError(t, os.RemoveAll(db_base_path))
}

func TestDbValidateKey(t *testing.T) {
	assert.NoError(t, dbValidateKey("a"))
	assert.NoError(t, dbValidateKey("a.json"))
	assert.NoError(t, dbValidateKey("a.txt"))
	assert.NoError(t, dbValidateKey("a.html"))
	assert.NoError(t, dbValidateKey("a/b"))
	assert.NoError(t, dbValidateKey("a//b"))
	assert.NoError(t, dbValidateKey("/a/b"))
	assert.NoError(t, dbValidateKey("a/b/c.d-E_F9"))

	assert.Error(t, dbValidateKey(""))
	assert.Error(t, dbValidateKey(" "))
	assert.Error(t, dbValidateKey("/"))
	assert.Error(t, dbValidateKey("a/"))
	assert.Error(t, dbValidateKey("a/b/"))
	assert.Error(t, dbValidateKey("a/b//"))
	assert.Error(t, dbValidateKey("a/b/ "))
	assert.Error(t, dbValidateKey("x?"))
	assert.Error(t, dbValidateKey("a b"))
	assert.Error(t, dbValidateKey("<!-->"))
	assert.Error(t, dbValidateKey("\\"))
}

func TestDbGetEntryPath(t *testing.T) {
	assert.Equal(t, "test/test_data/a", _dbGetEntryPath("a"))
	assert.Equal(t, "test/test_data/a.json", _dbGetEntryPath("/a.json"))
	assert.Equal(t, "test/test_data/a/b.json", _dbGetEntryPath("a/b.json"))
	assert.Equal(t, "test/test_data/a/b.json", _dbGetEntryPath("a//b.json"))
	assert.Equal(t, "test/test_data/a/b.json", _dbGetEntryPath("/a/b.json"))
}

func TestDbReadWrite(t *testing.T) {
	clearDataDir(t)

	assert.NoError(t, dbWriteEntry("x.json", []byte("[]")))
	assert.NoError(t, dbWriteEntry("/x.json", []byte("[2]")))
	assert.NoError(t, dbWriteEntry("/x/y.json", []byte("[3]")))

	assert.Error(t, dbWriteEntry("not_json.json", []byte("not json")))
	assert.NoError(t, dbWriteEntry("not_json.txt", []byte("not json")))

	checkRead := func(key string, expected []byte) {
		value, err := dbReadEntry(key)
		assert.NoError(t, err, "key %s", key)
		assert.Equal(t, expected, value, "key %s", key)
	}
	checkRead("x.json", []byte("[2]"))
	checkRead("/x.json", []byte("[2]"))
	checkRead("x/y.json", []byte("[3]"))
	checkRead("/x/y.json", []byte("[3]"))
	checkRead("/x//y.json", []byte("[3]"))

	checkReadFail := func(key string) {
		_, err := dbReadEntry(key)
		assert.Error(t, err, "key %s", key)
	}
	checkReadFail("a")
	checkReadFail("x.json/a")

	// invalid keys:
	checkReadFail("x/y/")
	checkReadFail("x/")
	checkReadFail("x.json/")
	checkReadFail("")
}

func TestDbReadPrefix(t *testing.T) {
	clearDataDir(t)

	assert.NoError(t, dbWriteEntry("a/1.txt", []byte("1")))
	assert.NoError(t, dbWriteEntry("a/2.json", []byte("[2]")))
	assert.NoError(t, dbWriteEntry("b/3.json", []byte("[3]")))

	checkList := func(list_fn func(string) ([]string, error), prefix string, expected_keys []string) {
		keys, err := list_fn(prefix)
		assert.NoError(t, err, "prefix %s", prefix)
		assert.Equal(t, expected_keys, keys, "prefix %s", prefix)
	}
	checkList(dbListPrefix, "a", []string{"a/1.txt", "a/2.json"})
	checkList(dbListPrefix, "/a", []string{"/a/1.txt", "/a/2.json"})
	checkList(dbListPrefix, "//a", []string{"/a/1.txt", "/a/2.json"})
	checkList(dbListPrefix, "b", []string{"b/3.json"})
	checkList(dbListSubprefixes, "a", []string{})
	checkList(dbListSubprefixes, "a/1.txt", []string{})

	checkReadAll := func(prefix string, expected map[string]interface{}) {
		results, errors, err := dbReadAllPrefix(prefix)
		assert.NoError(t, err)
		for key, expected_value := range expected {
			msg := fmt.Sprintf("prefix %s: key %s", prefix, key)
			switch expected_value.(type) {
			case []byte:
				assert.Equal(t, expected_value, results[key], msg)
				assert.NoError(t, errors[key], msg)
			case error:
				assert.Nil(t, results[key], msg)
				assert.Error(t, errors[key], msg)
			case nil:
				assert.Nil(t, results[key], msg)
				assert.Nil(t, errors[key], msg)
			default:
				assert.Fail(t, "invalid type %T: %s", expected_value, msg)
			}
		}
		for key, _ := range results {
			if _, ok := expected[key]; !ok {
				assert.Fail(t, "found unexpected key: prefix %s: %s", prefix, key)
			}
		}
	}
	assert.NoError(t, dbWriteEntry("a/invalid.txt", []byte("invalid")))
	os.Rename(_dbGetEntryPath("a/invalid.txt"), _dbGetEntryPath("a/invalid.json"))
	checkReadAll("a", map[string]interface{}{
		"a/1.txt":        []byte("1"),
		"a/2.json":       []byte("[2]"),
		"a/2":            nil,
		"a/invalid.json": errors.New(""),
	})
}
