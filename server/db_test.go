package main

import (
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
	assert.Equal(t, "test/test_data/a.json", dbGetEntryPath("a"))
	assert.Equal(t, "test/test_data/a.json", dbGetEntryPath("/a"))
	assert.Equal(t, "test/test_data/a/b.json", dbGetEntryPath("a/b"))
	assert.Equal(t, "test/test_data/a/b.json", dbGetEntryPath("a//b"))
	assert.Equal(t, "test/test_data/a/b.json", dbGetEntryPath("/a/b"))
}

func TestDbReadWrite(t *testing.T) {
	clearDataDir(t)

	assert.NoError(t, dbWriteEntry("x", []byte("[]")))
	assert.NoError(t, dbWriteEntry("/x", []byte("[2]")))
	assert.NoError(t, dbWriteEntry("/x/y", []byte("[3]")))

	checkRead := func(key string, expected []byte) {
		value, err := dbReadEntry(key)
		assert.NoError(t, err, "key %s", key)
		assert.Equal(t, expected, value, "key %s", key)
	}
	checkRead("x", []byte("[2]"))
	checkRead("/x", []byte("[2]"))
	checkRead("x/y", []byte("[3]"))
	checkRead("/x/y", []byte("[3]"))
	checkRead("/x//y", []byte("[3]"))

	checkReadFail := func(key string) {
		_, err := dbReadEntry(key)
		assert.Error(t, err, "key %s", key)
	}
	checkReadFail("a")
	checkReadFail("x/a")

	// invalid keys:
	checkReadFail("x/y/")
	checkReadFail("x/")
	checkReadFail("")
}
