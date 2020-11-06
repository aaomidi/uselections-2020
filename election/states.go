package election

import "sort"

var usc = map[string]string{
	"AZ": "Arizona",
	"GA": "Georgia",
	"ME": "Maine",
	"MI": "Michigan",
	"NC": "North Carolina",
	"NV": "Nevada",
	"PA": "Pennsylvania",
	"WI": "Wisconsin",
}

func StateExists(code string) bool {
	_, ok := usc[code]
	return ok
}

func GetStates() []State {
	states := GetIndexStates()
	v := make([]State, 0, len(states))

	for _, val := range states {
		v = append(v, val)
	}

	sort.Slice(v, func(i, j int) bool {
		return v[i].Name < v[j].Name
	})

	return v
}

func GetIndexStates() map[string]State {
	result := make(map[string]State)

	for key, value := range usc {
		result[key] = State{
			Name:         value,
			Abbreviation: key,
		}
	}

	return result
}
