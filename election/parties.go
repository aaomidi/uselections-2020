package election

import "strings"

var partyLookup = map[string]Party{
	"dem": {
		Name:         "Democrat",
		Symbol:       "🐴",
		Color:        "blue",
		Abbreviation: "Dem",
	},
	"gop": {
		Name:         "Republican",
		Symbol:       "🐘",
		Color:        "red",
		Abbreviation: "GOP",
	},
}

func GetParty(abbr string) Party {
	return partyLookup[strings.ToLower(abbr)]
}
