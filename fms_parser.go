package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"os"
	"strings"
)

// func main() {
// 	val := ParseHTMLtoJSON("raw0.html")
// 	fmt.Println(string(val))
// }

func split_and_strip(text string, separator string) ([]string) {
	// Split text into parts at separator character. Remove whitespace from parts.
	// "a • b • c" -> ["a", "b", "c"]
	parts := strings.Split(text, separator)
	var result []string
	for _, part := range parts {
		result = append(result, strings.TrimSpace(part))
	}
	return result
}

func ParseHTMLtoJSON(filename string) ([]byte) {
	//////////////////////////////////////////////////
	// Parse html from FMS into TBA-compatible JSON //
	//////////////////////////////////////////////////

	// Open file
	r, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file", filename)
		os.Exit(1)
	}

	// Read from file
	dom, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		fmt.Println("Error reading from file", filename)
		os.Exit(1)
	}

	// Parse file into map
	elements := make(map[string]map[string]string)
	elements["red"] = make(map[string]string)
	elements["blue"] = make(map[string]string)

	dom.Find("tr").Each(func(i int, s *goquery.Selection){
		columns := s.Children()
		if columns.Length() == 3 {
			var infos [3]string
			columns.Each(func(ii int, column *goquery.Selection){
				infos[ii] = column.Text()
			})
			identifier := infos[1]

			// Handle each data row
			if identifier == "" {
				// Skip
			} else if identifier == "Ownership Points" {
				key := "autoOwnershipPoints"
				if _, in := elements["red"][key]; in {
					key = "teleopOwnershipPoints"
				}
				elements["red"][key] = infos[0]
				elements["blue"][key] = infos[2]
			} else if identifier == "Auto-Run" {
				red_auto := split_and_strip(infos[0], "•")
				blue_auto := split_and_strip(infos[2], "•")
				elements["red"]["autoRobot1"] = red_auto[0]
				elements["red"]["autoRobot2"] = red_auto[1]
				elements["red"]["autoRobot3"] = red_auto[2]
				elements["blue"]["autoRobot1"] = blue_auto[0]
				elements["blue"]["autoRobot2"] = blue_auto[1]
				elements["blue"]["autoRobot3"] = blue_auto[2]
			} else if identifier == "Switch / Scale Ownership Seconds" {
				period := "auto"
				if _, in := elements["red"]["autoScaleOwnershipSec"]; in {
					period = "teleop"
				}
				red_ownership := split_and_strip(infos[0], "\n")
				blue_ownership := split_and_strip(infos[2], "\n")
				elements["red"][period + "SwitchOwnershipSec"] = red_ownership[0]
				elements["red"][period + "ScaleOwnershipSec"] = red_ownership[1]
				elements["blue"][period + "SwitchOwnershipSec"] = blue_ownership[0]
				elements["blue"][period + "ScaleOwnershipSec"] = blue_ownership[1]
			} else {
				elements["red"][identifier] = infos[0]
				elements["blue"][identifier] = infos[2]
			}
		}

	})

	res, err := json.Marshal(elements)
	if err != nil {
		fmt.Println("Failed to convert result to JSON:", err)
		os.Exit(1)
	}
	return res
}
