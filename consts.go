package main

const (
	MATCH_LEVEL_TEST = 0
	MATCH_LEVEL_PRACTICE = 1
	MATCH_LEVEL_QUAL = 2
	MATCH_LEVEL_PLAYOFF = 3
)

func getConstsMap() map[string]interface{} {
	m := make(map[string]interface{})

	m["MATCH_LEVEL_TEST"]     = MATCH_LEVEL_TEST
	m["MATCH_LEVEL_PRACTICE"] = MATCH_LEVEL_PRACTICE
	m["MATCH_LEVEL_QUAL"]     = MATCH_LEVEL_QUAL
	m["MATCH_LEVEL_PLAYOFF"]  = MATCH_LEVEL_PLAYOFF

	return m
}
