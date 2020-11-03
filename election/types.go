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

func (p *Party) Single() string {
	return string(p.Name[0])
}

type State struct {
	Name         string
	Abbreviation string
}

// Vote is the representation for the votes of a election in a given state
type Vote struct {
	Candidate      Candidate
	State          State
	Count          int64
	Percentage     float64
	ElectoralVotes int
	StateVote      StateResults // Link to the information about the entire state
}

// StateResults is the representation of the state of voting in a given state
type StateResults struct {
	State               State
	TotalVotes          int64
	TurnoutPercentage   float64
	ReportingPercentage float64
	ReportingCount      int
	TotalPrecincts      int
	Winner              []Winner
}

// Winner represents a candidate and their electoral votes
type Winner struct {
	Candidate      Candidate
	ElectoralVotes int
}
