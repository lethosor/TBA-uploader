package fms_parser

import (
	"testing"
)

func TestParse2023(t *testing.T) {
	testParseMatchDir(t, parseHTMLtoJSON2023, "../tests/data/2023/")
}
