package scraper

import (
	"context"
	"github.com/aaomidi/uselections-2020/election"
)

type Scraper interface {
	Scrape(context context.Context) <-chan election.Vote
}
