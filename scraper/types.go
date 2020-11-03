package scraper

import "github.com/aaomidi/uselections-2020/election"

type Scraper interface {
	Scrape() []election.Vote
}
