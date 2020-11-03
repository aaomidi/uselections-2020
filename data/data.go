package data

import (
	"context"
	"github.com/aaomidi/uselections-2020/election"
	"github.com/aaomidi/uselections-2020/scraper"
	log "github.com/sirupsen/logrus"
	"time"
)

type Data struct {
	broadcaster chan<- BroadcastRequest
}

type BroadcastRequest struct {
	// listenerWritable will accept writable channels to broadcast vote results
	listenerWritable chan<- OutgoingUpdate
}

type OutgoingUpdate struct {
	// Votes contains all the updated votes
	Votes []election.Vote

	// NotificationVotes these are the interesting votes that passed a certain threshold and we should tell users about that
	NotificationVotes []election.Vote
}

func (d *Data) Start(s scraper.Scraper) {
	broadcaster := make(chan BroadcastRequest)
	d.broadcaster = broadcaster

	aggregation := make(chan []election.Vote)
	go func(s scraper.Scraper) {

		for range time.Tick(5 * time.Second) {
			votes := make([]election.Vote, 0, 153)
			for vote := range s.Scrape(context.Background()) {
				votes = append(votes, vote)
			}

			aggregation <- votes
		}
	}(s)

	go d.aggregate(aggregation, broadcaster)
}

func (d *Data) aggregate(incoming <-chan []election.Vote, broadcastRequests <-chan BroadcastRequest) {
	listeners := make([]BroadcastRequest, 0, 5)
	for {
		select {
		case newBroadcast := <-broadcastRequests:
			listeners = append(listeners, newBroadcast)
		case newVoteBucket := <-incoming:
			for _, listener := range listeners {
				select {
				case listener.listenerWritable <- OutgoingUpdate{
					Votes:             newVoteBucket,
					NotificationVotes: nil,
				}:
				default:
					log.Warning("Some listener was full :/")
				}
			}
		}
	}
}

func (d *Data) RegisterDataReceiver(writable chan<- OutgoingUpdate) {
	d.broadcaster <- BroadcastRequest{
		listenerWritable: writable,
	}
}
