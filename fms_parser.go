package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

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

type fmsScoreInfo struct {
	auto int64
	teleop int64
	fouls int64
	total int64

	autoRunPoints int64
	autoSwitchSec int64
}

func ParseHTMLtoJSON(filename string) (map[string]interface{}, error) {
	//////////////////////////////////////////////////
	// Parse html from FMS into TBA-compatible JSON //
	//////////////////////////////////////////////////

	// Open file
	r, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Error opening file: %s", filename)
	}

	// Read from file
	dom, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("Error reading from file: %s", filename)
	}

	all_json := make(map[string]interface{})

	alliances := map[string]map[string]interface{} {
		"blue": map[string]interface{} {
			"teams": make([]string, 3),
			"surrogates": make([]string, 0),
			"dqs": make([]string, 0),
			"score": -1,
		},
		"red": map[string]interface{} {
			"teams": make([]string, 3),
			"surrogates": make([]string, 0),
			"dqs": make([]string, 0),
			"score": -1,
		},
	}

	breakdown := map[string]map[string]interface{} {
		"blue": make(map[string]interface{}),
		"red": make(map[string]interface{}),
	}

	var scoreInfo struct {
		blue fmsScoreInfo
		red fmsScoreInfo
	}

	parse_error := ""
	dom.Find("tr").Each(func(i int, s *goquery.Selection){
		columns := s.Children()
		if columns.Length() == 3 {
			var infos [3]string
			columns.Each(func(ii int, column *goquery.Selection){
				infos[ii] = strings.TrimSpace(column.Text())
			})
			identifier := infos[1]

			// Handle each data row
			if identifier == "" {
				// Skip
			} else if identifier == "Final Score" {
				blue_score, err := strconv.ParseInt(infos[0], 10, 0)
				red_score, err := strconv.ParseInt(infos[2], 10, 0)
				if err != nil {
					parse_error = "final score failed"
				}
				breakdown["blue"]["totalPoints"] = blue_score
				breakdown["red"]["totalPoints"] = red_score
				alliances["blue"]["score"] = blue_score
				alliances["red"]["score"] = red_score
				scoreInfo.blue.total = blue_score
				scoreInfo.red.total = red_score
			} else if identifier == "Teams" {
				blue_teams := split_and_strip(infos[0], "•")
				red_teams := split_and_strip(infos[2], "•")
				alliances["blue"]["teams"] = []string{
					"frc" + blue_teams[0],
					"frc" + blue_teams[1],
					"frc" + blue_teams[2],
				}
				alliances["red"]["teams"] = []string{
					"frc" + red_teams[0],
					"frc" + red_teams[1],
					"frc" + red_teams[2],
				}
			} else if identifier == "Ownership Points" {
				key := "autoOwnershipPoints"
				if _, in := breakdown["blue"][key]; in {
					key = "teleopOwnershipPoints"
				}
				breakdown["blue"][key], err = strconv.ParseInt(infos[0], 10, 0)
				breakdown["red"][key], err = strconv.ParseInt(infos[2], 10, 0)
				if err != nil {
					parse_error = key + " ownership points failed"
				}
			} else if identifier == "Auto-Run" {
				blue_auto := split_and_strip(infos[0], "•")
				red_auto := split_and_strip(infos[2], "•")
				breakdown["blue"]["autoRobot1"] = blue_auto[0]
				breakdown["blue"]["autoRobot2"] = blue_auto[1]
				breakdown["blue"]["autoRobot3"] = blue_auto[2]
				breakdown["red"]["autoRobot1"] = red_auto[0]
				breakdown["red"]["autoRobot2"] = red_auto[1]
				breakdown["red"]["autoRobot3"] = red_auto[2]
			} else if identifier == "Auto-Run Points" {
				blue_autorun_points, err := strconv.ParseInt(infos[0], 10, 0)
				red_autorun_points, err := strconv.ParseInt(infos[2], 10, 0)
				if err != nil {
					parse_error = "auto-run points failed"
				}
				scoreInfo.blue.autoRunPoints = blue_autorun_points
				breakdown["blue"]["autoRunPoints"] = blue_autorun_points
				scoreInfo.red.autoRunPoints = red_autorun_points
				breakdown["red"]["autoRunPoints"] = red_autorun_points
			} else if identifier == "Autonomous" {
				blue_auto_points, err := strconv.ParseInt(infos[0], 10, 0)
				red_auto_points, err := strconv.ParseInt(infos[2], 10, 0)
				if err != nil {
					parse_error = "autonomous points failed"
				}
				scoreInfo.blue.auto = blue_auto_points
				breakdown["blue"]["autoPoints"] = blue_auto_points
				scoreInfo.red.auto = red_auto_points
				breakdown["red"]["autoPoints"] = red_auto_points
			} else if identifier == "Switch / Scale Ownership Seconds" {
				period := "auto"
				if _, in := breakdown["blue"]["autoScaleOwnershipSec"]; in {
					period = "teleop"
				}
				blue_ownership := split_and_strip(infos[0], "\n")
				red_ownership := split_and_strip(infos[2], "\n")
				breakdown["blue"][period + "SwitchOwnershipSec"], err = strconv.ParseInt(blue_ownership[0], 10, 0)
				breakdown["blue"][period + "ScaleOwnershipSec"], err = strconv.ParseInt(blue_ownership[1], 10, 0)
				breakdown["red"][period + "SwitchOwnershipSec"], err = strconv.ParseInt(red_ownership[0], 10, 0)
				breakdown["red"][period + "ScaleOwnershipSec"], err = strconv.ParseInt(red_ownership[1], 10, 0)
				if period == "auto" {
					scoreInfo.blue.autoSwitchSec, err = strconv.ParseInt(blue_ownership[0], 10, 0)
					scoreInfo.red.autoSwitchSec, err = strconv.ParseInt(red_ownership[0], 10, 0)
				}
				if err != nil {
					parse_error = period + " ownership seconds failed"
				}
			} else if identifier == "Switch / Scale Boost Seconds" {
				blue_boost := split_and_strip(infos[0], "\n")
				red_boost := split_and_strip(infos[2], "\n")
				breakdown["blue"]["teleopSwitchBoostSec"], err = strconv.ParseInt(blue_boost[0], 10, 0)
				breakdown["blue"]["teleopScaleBoostSec"], err = strconv.ParseInt(blue_boost[1], 10, 0)
				breakdown["red"]["teleopSwitchBoostSec"], err = strconv.ParseInt(red_boost[0], 10, 0)
				breakdown["red"]["teleopScaleBoostSec"], err = strconv.ParseInt(red_boost[1], 10, 0)
				if err != nil {
					parse_error = "teleop boost seconds failed"
				}
			} else if identifier == "Switch / Scale Force Seconds" {
				blue_force := split_and_strip(infos[0], "\n")
				red_force := split_and_strip(infos[2], "\n")
				breakdown["blue"]["teleopSwitchForceSec"], err = strconv.ParseInt(blue_force[0], 10, 0)
				breakdown["blue"]["teleopScaleForceSec"], err = strconv.ParseInt(blue_force[1], 10, 0)
				breakdown["red"]["teleopSwitchForceSec"], err = strconv.ParseInt(red_force[0], 10, 0)
				breakdown["red"]["teleopScaleForceSec"], err = strconv.ParseInt(red_force[1], 10, 0)
				if err != nil {
					parse_error = "teleop force seconds failed"
				}
			} else if identifier == "Vault Points" {
				breakdown["blue"]["vaultPoints"], err = strconv.ParseInt(infos[0], 10, 0)
				breakdown["red"]["vaultPoints"], err = strconv.ParseInt(infos[2], 10, 0)
				if err != nil {
					parse_error = "vault points failed"
				}
			} else if identifier == "Endgame" {
				blue_endgame := split_and_strip(infos[0], "•")
				red_endgame := split_and_strip(infos[2], "•")
				breakdown["blue"]["endgameRobot1"] = blue_endgame[0]
				breakdown["blue"]["endgameRobot2"] = blue_endgame[1]
				breakdown["blue"]["endgameRobot3"] = blue_endgame[2]
				breakdown["red"]["endgameRobot1"] = red_endgame[0]
				breakdown["red"]["endgameRobot2"] = red_endgame[1]
				breakdown["red"]["endgameRobot3"] = red_endgame[2]
			} else if identifier == "Endgame Points" {
				breakdown["blue"]["endgamePoints"], err = strconv.ParseInt(infos[0], 10, 0)
				breakdown["red"]["endgamePoints"], err = strconv.ParseInt(infos[2], 10, 0)
				if err != nil {
					parse_error = "endgame points failed"
				}
			} else if identifier == "Teleop" {
				blue_teleop_points, err := strconv.ParseInt(infos[0], 10, 0)
				red_teleop_points, err := strconv.ParseInt(infos[2], 10, 0)
				if err != nil {
					parse_error = "teleop points failed"
				}
				breakdown["blue"]["teleopPoints"] = blue_teleop_points
				breakdown["red"]["teleopPoints"] = red_teleop_points
				scoreInfo.blue.teleop = blue_teleop_points
				scoreInfo.red.teleop = red_teleop_points
			} else if identifier == "Fouls/Techs Committed" {
				blue_foul := split_and_strip(infos[0], "•")
				red_foul := split_and_strip(infos[2], "•")
				breakdown["blue"]["foulCount"], err = strconv.ParseInt(blue_foul[0], 10, 0)
				breakdown["blue"]["techFoulCount"], err = strconv.ParseInt(blue_foul[1], 10, 0)
				breakdown["red"]["foulCount"], err = strconv.ParseInt(red_foul[0], 10, 0)
				breakdown["red"]["techFoulCount"], err = strconv.ParseInt(red_foul[1], 10, 0)
				if err != nil {
					parse_error = "foul/tech count failed"
				}
			} else if identifier == "Foul Points" {
				blue_foul_points, err := strconv.ParseInt(infos[0], 10, 0)
				red_foul_points, err := strconv.ParseInt(infos[2], 10, 0)
				if err != nil {
					parse_error = "foul points failed"
				}
				breakdown["blue"]["foulPoints"] = blue_foul_points
				breakdown["red"]["foulPoints"] = red_foul_points
				scoreInfo.blue.fouls = blue_foul_points
				scoreInfo.red.fouls = red_foul_points
			} else if identifier == "Force Powerup" || identifier == "Boost Powerup" {
				powerup := string(identifier[0:5])
				// first character
				blue_total, err := strconv.ParseInt(string(infos[0][0]), 10, 0)
				red_total, err := strconv.ParseInt(string(infos[2][0]), 10, 0)
				if err != nil {
					parse_error = "failed to parse powerup total: " + powerup
				}
				var blue_played int64
				var red_played int64
				if blue_total == 0 {
					blue_played = 0
				} else {
					if strings.HasSuffix(infos[0], "Not Played") {
						blue_played = 0
					} else {
						blue_played, err = strconv.ParseInt(string(infos[0][len(infos[0])-1:]), 10, 0)
						if err != nil {
							parse_error = "failed to parse powerup played: " + powerup + " (blue)"
						}
					}
				}
				if red_total == 0 {
					red_played = 0
				} else {
					if strings.HasSuffix(infos[2], "Not Played") {
						red_played = 0
					} else {
						red_played, err = strconv.ParseInt(string(infos[2][len(infos[2])-1:]), 10, 0)
						if err != nil {
							parse_error = "failed to parse powerup played: " + powerup + " (red)"
						}
					}
				}
				breakdown["blue"]["vault" + powerup + "Total"] = blue_total
				breakdown["blue"]["vault" + powerup + "Played"] = blue_played

				breakdown["red"]["vault" + powerup + "Total"] = red_total
				breakdown["red"]["vault" + powerup + "Played"] = red_played
			} else if identifier == "Levitate Powerup" {
				// first character
				blue_total, err := strconv.ParseInt(string(infos[0][0]), 10, 0)
				red_total, err := strconv.ParseInt(string(infos[2][0]), 10, 0)
				if err != nil {
					parse_error = "failed to parse levitate total"
				}
				blue_played := 0
				red_played := 0
				if blue_total == 3 && strings.HasSuffix(infos[0], ", Played") {
					blue_played = 3
				}
				if red_total == 3 && strings.HasSuffix(infos[2], ", Played") {
					red_played = 3
				}
				breakdown["blue"]["vaultLevitateTotal"] = blue_total
				breakdown["blue"]["vaultLevitatePlayed"] = blue_played
				breakdown["red"]["vaultLevitateTotal"] = red_total
				breakdown["red"]["vaultLevitatePlayed"] = red_played
			} else {
				breakdown["blue"][identifier] = strings.TrimSpace(infos[0])
				breakdown["red"][identifier] = strings.TrimSpace(infos[2])
			}
		}
	})

	gamedata := dom.Find(".panel-body.text-center").Text()
	breakdown["blue"]["tba_gameData"] = gamedata
	breakdown["red"]["tba_gameData"] = gamedata

	if parse_error != "" {
		return nil, fmt.Errorf("Parse error: %s", parse_error)
	}

	all_json["alliances"] = alliances
	all_json["score_breakdown"] = breakdown

	return all_json, nil
}
