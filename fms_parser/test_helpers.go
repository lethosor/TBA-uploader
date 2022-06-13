package fms_parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-test/deep"
)

type testTbaAllianceTeams struct {
	TeamKeys []string `json:"teams" json:"team_keys"`
}

type testTbaMatchResult struct {
	CompLevel string `json:"comp_level"`
	Alliances struct {
		Blue map[string]interface{} `json:"blue"`
		Red  map[string]interface{} `json:"red"`
	} `json:"alliances"`
	ScoreBreakdown struct {
		Blue map[string]interface{} `json:"blue"`
		Red  map[string]interface{} `json:"red"`
	} `json:"score_breakdown"`
}

func testParseSingleMatch(
	t *testing.T,
	parser func(filename string, playoff bool) (map[string]interface{}, error),
	fms_html_path string,
	tba_json_path string,
) {
	json_contents, err := ioutil.ReadFile(tba_json_path)
	if err != nil {
		t.Errorf("%s: %s", fms_html_path, err)
		return
	}

	tba_result := testTbaMatchResult{}
	json.Unmarshal(json_contents, &tba_result)

	is_playoff := (tba_result.CompLevel != "qm")
	parsed_json, err := parser(fms_html_path, is_playoff)
	if err != nil {
		t.Errorf("%s: %s", fms_html_path, err)
		return
	}

	parsed_marshaled, _ := json.MarshalIndent(parsed_json, "", "  ")
	parsed_result := testTbaMatchResult{}
	json.Unmarshal(parsed_marshaled, &parsed_result);

	if os.Getenv("TEST_DEBUG") != "" {
		fmt.Printf("%s", parsed_marshaled)
	}

	deep.MaxDiff = 100

	// write API expects "teams", but read API returns "team_keys"
	if diff := deep.Equal(parsed_result.Alliances.Blue["teams"], tba_result.Alliances.Blue["team_keys"]); diff != nil {
		t.Errorf("%s: blue alliance does not match: %s", fms_html_path, diff)
	}
	if diff := deep.Equal(parsed_result.Alliances.Red["teams"], tba_result.Alliances.Red["team_keys"]); diff != nil {
		t.Errorf("%s: red alliance does not match: %s", fms_html_path, diff)
	}

	if diff := deep.Equal(parsed_result.ScoreBreakdown, tba_result.ScoreBreakdown); diff != nil {
		t.Errorf("%s: breakdown does not match: %s", fms_html_path, diff)
	}
}

func testParseMatchDir(
	t *testing.T,
	parser func(filename string, playoff bool) (map[string]interface{}, error),
	dirname string,
) {
	all_files, err := ioutil.ReadDir(dirname)
	if err != nil {
	    t.Error(err)
	    return

	}
	for _, file := range all_files {
	    if file.Mode().IsRegular() && filepath.Ext(file.Name()) == ".html" {
	        testParseSingleMatch(
	        	t,
	        	parser,
	        	path.Join(dirname, file.Name()),
	        	path.Join(dirname, strings.Replace(file.Name(), ".html", ".json", 1)),
	        )
	    }
	}
}
