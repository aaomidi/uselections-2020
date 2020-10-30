package election

type Candidate struct {
	FirstName string
	LastName  string
	Party     Party
}

type Party struct {
	Name         string
	Symbol       string
	Color        string
	Abbreviation string
}

type State struct {
	Name         string
	Abbreviation string
}

// Vote is the representation for the votes of a election in a given state
type Vote struct {
	Candidate      Candidate
	State          State
	Count          int
	Percentage     int
	ElectoralVotes int
	StateVote      *StateVote // Link to the information about the entire state
}

// StateVote is the representation of the state of voting in a given state
type StateVote struct {
	State               State
	TotalVotes          int
	TurnoutPercentage   float64
	ReportingPercentage float64
	ReportingCount      int
	TotalPrecincts      int
}
