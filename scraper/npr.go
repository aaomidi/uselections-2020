package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aaomidi/uselections-2020/election"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	//AllStatesURL shows every state
	AllStatesURL = "https://apps.npr.org/elections20-interactive/data/president.json"

	//StateFormatURL is used to format the URL
	StateFormatURL = "https://apps.npr.org/elections20-interactive/data/states/%s.json"
)

type NPRStateData struct {
	Results []NPRElectionData
}

//Transform transforms the NPRData to our data format
func (data *NPRStateData) Transform() []election.Vote {
	var votes []election.Vote

	for _, result := range data.Results {
		// currently we only care about president. NPR offers senate. If we have time (which we likely won't)
		// We can add senate and house since the data is easily accessible.
		if !result.Test && result.Office == "P" {
			votes = result.Transform(votes)
		}
	}

	return votes
}

type NPRElectionData struct {
	// If this is test data - hopefully we won't come across this in production. that'd be embarrassing
	Test bool

	// The office race - "P" for president, "S" for senate, "G" for Governor, "H" for House
	Office string

	// The type of election. Some states are having special elections to fill vacancies. "special", "general"
	Type string

	// The level we're viewing. Most of the time this will be "state"
	// But there are a total of 7 instances where it will show the districts
	// Maine's 2 congressional districts + an At Large
	// and Nebraska's 3 Congressional districts + an At Large
	// Now. I have no idea why NPR includes the At Large when those seats were eliminated
	// in the late 1800s. But you do you NPR
	Level string

	// The state code
	State string

	// The name of the state
	StateName string

	// The states abbreviation
	StateAP string

	// The district - only used in 2 states
	District string

	// The total number of precincts
	Precincts int

	// The timestamp of when the information was last updated
	Updated int64

	// The total number of precincts reporting
	Reporting int

	// The percentage of precincts reporting
	ReportingPercent float64 `json:",omitempty"`

	// The number of electoral votes up for grabs in this state
	Electoral int

	// The candidates in this states ballot
	Candidates []NPRCandidateData
}

func (data *NPRElectionData) Transform(votes []election.Vote) []election.Vote {
	stateResults := election.StateResults{
		State: election.State{
			Name:         data.StateName,
			Abbreviation: data.State,
		},
		ReportingCount:      data.Reporting,
		ReportingPercentage: data.ReportingPercent,
		TotalPrecincts:      data.Precincts,
	}

	//var votes []election.Vote

	for _, candidate := range data.Candidates {
		stateResults.TotalVotes += candidate.Votes

		vote := election.Vote{
			Candidate: election.Candidate{
				FirstName: candidate.First,
				LastName:  candidate.Last,
				Party:     election.GetParty(candidate.Party),
			},
			Percentage:     candidate.Percent,
			Count:          candidate.Votes,
			ElectoralVotes: candidate.Electoral,
			StateVote:      stateResults,
		}

		votes = append(votes, vote)
	}

	return votes
}

type NPRCandidateData struct {
	// The first and last names of the candidates on the ballot
	// NPR doesn't list the third party candidates from what it seems so it's listed as "other"
	First, Last string

	// The party of the candidate. "GOP" - Republican, "Dem" - Democrat
	Party string

	// The total number of votes a candidate has
	Votes int64

	// I have no idea what this field is for. I assume absentee ballots?
	AVotes int64

	// Is the candidate incumbent?
	Incumbent bool

	// The number of electoral votes won by the candidate for this state/district
	Electoral int

	Percent float64 `json:",omitempty"`

	// This is the number of third party candidates on the ballot
	Count int `json:",omitempty"`
}

func FormatURL(code string) string {
	code = strings.ToUpper(code)

	if election.HasState(code) {
		return fmt.Sprintf(StateFormatURL, code)
	}
	return AllStatesURL
}

//NPRScraper is an implementation of the Scraper interface
//using the NPR interactive election data
// URL: https://apps.npr.org/elections20-interactive/data/president.json
type NPRScraper struct{}

func (npr *NPRScraper) Scrape(ctx context.Context) <-chan election.Vote {
	channel := make(chan election.Vote)

	go func(ctx context.Context) {
		state := npr.getStateFromContext(ctx)

		results, err := npr.Fetch(state)

		if err != nil {
			// not sure what to do here ngl. it's 3am and i'm too tired. do we cancel the context?
			panic(err)
		}

		for _, vote := range results {
			channel <- vote
		}
	}(ctx)

	return channel
}

func (npr *NPRScraper) getStateFromContext(ctx context.Context) string {
	state := ctx.Value("state")

	if state == nil {
		return ""
	}

	return state.(string)
}

func (npr *NPRScraper) Fetch(state string) ([]election.Vote, error) {
	response, err := http.Get(FormatURL(state))

	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	var nprData = &NPRStateData{}

	err = json.Unmarshal(data, nprData)

	if err != nil {
		return nil, err
	}

	transformed := nprData.Transform()

	return transformed, nil
}
