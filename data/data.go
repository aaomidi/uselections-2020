package data

import (
	"context"
	"github.com/aaomidi/uselections-2020/election"
	"github.com/aaomidi/uselections-2020/scraper"
	"time"
)

type Data struct {
	broadcaster chan<- BroadcastRequest
}

type BroadcastRequest struct {
	// listenerWritable will accept writable channels to broadcast vote results
	listenerWritable chan<- []election.Vote
}

func (d *Data) Start(s scraper.Scraper) {
	broadcaster := make(chan BroadcastRequest)
	d.broadcaster = broadcaster

	aggregation := make(chan []election.Vote, 1)
	go func(s scraper.Scraper) {

		for range time.Tick(5 * time.Second) {
			votes := make([]election.Vote, 0, 50)
			for vote := range s.Scrape(context.Background()) {
				votes = append(votes, vote)
			}

			aggregation <- votes
		}
	}(s)

	d.aggregate(aggregation, broadcaster)
}

func (d *Data) aggregate(incoming <-chan []election.Vote, broadcastRequests <-chan BroadcastRequest) {
	go func() {
		listeners := make([]BroadcastRequest, 0, 5)
		for {
			select {
			case newBroadcast := <-broadcastRequests:
				listeners = append(listeners, newBroadcast)
			case newVoteBucket := <-incoming:
				for _, listener := range listeners {
					listener.listenerWritable <- newVoteBucket
				}
			}
		}
	}()
}

func (d *Data) RegisterDataReceiver(writable chan<- []election.Vote) {
	d.broadcaster <- BroadcastRequest{
		listenerWritable: writable,
	}
}
