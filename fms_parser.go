package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"os"
	"strconv"
	"strings"
)

func main() {
	val, _ := ParseHTMLtoJSON("raw0.html")
	fmt.Println(string(val))
}

func split_and_strip(text string, separator string) ([]string) {
	// Split text into parts at separator character. Remove whitespace from parts.
	// "a • b • c" -> ["a", "b", "c"]
	parts := strings.Split(strings.TrimSpace(text), separator)
	var result []string
	for _, part := range parts {
		result = append(result, strings.TrimSpace(part))
	}
	return result
}

func ParseHTMLtoJSON(filename string) ([]byte, error) {
	//////////////////////////////////////////////////
	// Parse html from FMS into TBA-compatible JSON //
	//////////////////////////////////////////////////

	// Open file
	r, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file", filename)
		return nil, errors.New("Error")
	}

	// Read from file
	dom, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		fmt.Println("Error reading from file", filename)
		return nil, errors.New("Error")
	}

	// Parse file into map
	elements := make(map[string]map[string]string)
	elements["blue"] = make(map[string]string)
	elements["red"] = make(map[string]string)

	parse_error := ""
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
			} else if identifier == "Teams"{
				blue_teams := split_and_strip(infos[0], "•")
				red_teams := split_and_strip(infos[2], "•")
				elements["blue"]["team1"] = blue_teams[0]
				elements["blue"]["team2"] = blue_teams[1]
				elements["blue"]["team3"] = blue_teams[2]
				elements["red"]["team1"] = red_teams[0]
				elements["red"]["team2"] = red_teams[1]
				elements["red"]["team3"] = red_teams[2]
			} else if identifier == "Ownership Points" {
				key := "autoOwnershipPoints"
				if _, in := elements["blue"][key]; in {
					key = "teleopOwnershipPoints"
				}
				elements["blue"][key] = infos[0]
				elements["red"][key] = infos[2]
			} else if identifier == "Auto-Run" {
				blue_auto := split_and_strip(infos[0], "•")
				red_auto := split_and_strip(infos[2], "•")
				elements["blue"]["autoRobot1"] = blue_auto[0]
				elements["blue"]["autoRobot2"] = blue_auto[1]
				elements["blue"]["autoRobot3"] = blue_auto[2]
				elements["red"]["autoRobot1"] = red_auto[0]
				elements["red"]["autoRobot2"] = red_auto[1]
				elements["red"]["autoRobot3"] = red_auto[2]
			} else if identifier == "Switch / Scale Ownership Seconds" {
				period := "auto"
				if _, in := elements["blue"]["autoScaleOwnershipSec"]; in {
					period = "teleop"
				}
				blue_ownership := split_and_strip(infos[0], "\n")
				red_ownership := split_and_strip(infos[2], "\n")
				elements["blue"][period + "SwitchOwnershipSec"] = blue_ownership[0]
				elements["blue"][period + "ScaleOwnershipSec"] = blue_ownership[1]
				elements["red"][period + "SwitchOwnershipSec"] = red_ownership[0]
				elements["red"][period + "ScaleOwnershipSec"] = red_ownership[1]
			} else if identifier == "Switch / Scale Boost Seconds" {
				blue_boost := split_and_strip(infos[0], "\n")
				red_boost := split_and_strip(infos[2], "\n")
				elements["blue"]["teleopSwitchBoostSec"] = blue_boost[0]
				elements["blue"]["teleopScaleBoostSec"] = blue_boost[1]
				elements["red"]["teleopSwitchBoostSec"] = red_boost[0]
				elements["red"]["teleopScaleBoostSec"] = red_boost[1]
			} else if identifier == "Switch / Scale Force Seconds" {
				blue_force := split_and_strip(infos[0], " ")
				red_force := split_and_strip(infos[2], " ")
				elements["blue"]["teleopSwitchForceSec"] = blue_force[0]
				elements["blue"]["teleopScaleForceSec"] = blue_force[1]
				elements["red"]["teleopSwitchForceSec"] = red_force[0]
				elements["red"]["teleopScaleForceSec"] = red_force[1]
			} else if identifier == "Endgame" {
				blue_endgame := split_and_strip(infos[0], "•")
				red_endgame := split_and_strip(infos[2], "•")
				elements["blue"]["endgameRobot1"] = blue_endgame[0]
				elements["blue"]["endgameRobot2"] = blue_endgame[1]
				elements["blue"]["endgameRobot3"] = blue_endgame[2]
				elements["red"]["endgameRobot1"] = red_endgame[0]
				elements["red"]["endgameRobot2"] = red_endgame[1]
				elements["red"]["endgameRobot3"] = red_endgame[2]
			} else if identifier == "Fouls/Techs Committed" {
				blue_foul := split_and_strip(infos[0], "•")
				red_foul := split_and_strip(infos[2], "•")
				elements["blue"]["foulCount"] = blue_foul[0]
				elements["blue"]["techFoulCount"] = blue_foul[1]
				elements["red"]["foulCount"] = red_foul[0]
				elements["red"]["techFoulCount"] = red_foul[1]
			} else if identifier == "Force Powerup" || identifier == "Boost Powerup" {
				powerup := string(identifier[0:5])
				blue_total := string(infos[0][0]) // First character
				red_total := string(infos[2][0])
				var blue_played string
				var red_played string
				if val, err := strconv.Atoi(blue_total); err != nil {
					parse_error = "Error"
				} else if val == 0 {
					blue_played = "0"
				} else {
					if strings.HasSuffix(infos[0], "Not Played") {
						blue_played = "0"
					} else {
						blue_played = string(infos[0][len(infos[0])-1:])
					}
				}
				if val, err := strconv.Atoi(red_total); err != nil {
					parse_error = "Error"
				} else if val == 0 {
					red_played = "0"
				} else {
					if strings.HasSuffix(infos[2], "Not Played") {
						red_played = "0"
					} else {
						red_played = string(infos[2][len(infos[2])-1:])
					}
				}
				elements["blue"]["vault" + powerup + "Total"] = blue_total
				elements["blue"]["vault" + powerup + "Played"] = blue_played

				elements["red"]["vault" + powerup + "Total"] = red_total
				elements["red"]["vault" + powerup + "Played"] = red_played
			} else if identifier == "Levitate Powerup" {
				blue_total := string(infos[0][0]) // First character
				red_total := string(infos[2][0])
				blue_played := "0"
				red_played := "0"
				if val, err := strconv.Atoi(blue_total); err != nil {
					parse_error = "Error"
				} else if val == 3 && strings.HasSuffix(infos[0], "Played") {
					blue_played = "3"
				}
				if val, err := strconv.Atoi(red_total); err != nil {
					parse_error = "Error"
				} else if val == 3 && strings.HasSuffix(infos[0], "Played") {
					red_played = "3"
				}
				elements["blue"]["vaultLevitateTotal"] = blue_total
				elements["blue"]["vaultLevitatePlayed"] = blue_played
				elements["red"]["vaultLevitateTotal"] = red_total
				elements["red"]["vaultLevitatePlayed"] = red_played
			} else {
				elements["blue"][identifier] = strings.TrimSpace(infos[0])
				elements["red"][identifier] = strings.TrimSpace(infos[2])
			}
		}
	})
	if parse_error != "" {
		return nil, errors.New(parse_error)
	}

	res, err := json.Marshal(elements)
	if err != nil {
		fmt.Println("Failed to convert result to JSON:", err)
		return nil, errors.New("Error")
	}
	return res, nil
}
