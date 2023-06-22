package fms_parser

import (
	"testing"
)

func TestParse2022(t *testing.T) {
	testParseMatchDir(t, parseHTMLtoJSON2022, "../tests/data/2022/")
}
