package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsFile(t *testing.T) {
	assert.True(t, isFile("main.go"))

	assert.False(t, isFile("asdf.go"))
}

func TestIsDir(t *testing.T) {
	assert.True(t, isDir("."))
	assert.True(t, isDir(".."))

	assert.False(t, isDir("asdf"))
}
