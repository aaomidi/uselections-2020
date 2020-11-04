package telegram

import (
	"fmt"
	"github.com/aaomidi/uselections-2020/data"
	"github.com/aaomidi/uselections-2020/election"
	"github.com/aaomidi/uselections-2020/redis"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
	"strconv"
	"strings"
	"time"
)

type Telegram struct {
	token       string
	bot         *tb.Bot
	channelID   string
	channel     *tb.Chat
	log         *log.Entry
	redis       *redis.Redis
	data        *data.Data
	dataChannel chan data.OutgoingUpdate
}

func New(token string, channelID string, r *redis.Redis, d *data.Data) Telegram {
	return Telegram{
		token:     token,
		bot:       nil,
		channelID: channelID,
		log:       log.WithField("source", "telegram"),
		redis:     r,
		data:      d,
	}
}

func (t *Telegram) Create() error {
	bot, err := tb.NewBot(tb.Settings{
		Token:   t.token,
		Poller:  &tb.LongPoller{Timeout: 15 * time.Second},
		Verbose: true,
	})
	t.bot = bot

	if err != nil {
		return NewError(err, "creation failed")
	}

	channel, err := t.bot.ChatByID(t.channelID)

	if err != nil {
		return NewError(err, "chat id did not resolve :(")
	}

	t.channel = channel

	t.log.Infof("connected to %s", t.channelID)

	return nil
}

func (t *Telegram) Start() {
	t.dataChannel = make(chan data.OutgoingUpdate)

	t.data.RegisterDataReceiver(t.dataChannel)

	go t.runListener()

	// On inline query
	t.bot.Handle(tb.OnQuery, t.handleQuery)

	// On inline chosen result
	t.bot.Handle(tb.OnChosenInlineResult, t.handleChosenInlineResult)

	// Start the bot, listen for queries
	t.bot.Start()
}

func (t *Telegram) handleQuery(q *tb.Query) {
	msg := q.Text

	if len(msg) != 2 {
		_ = t.bot.Answer(q, &tb.QueryResponse{
			Results:   nil,
			CacheTime: 60,
		})
		return
	}
	states := election.GetIndexStates()

	state, ok := states[strings.ToUpper(msg)]

	if !ok {
		_ = t.bot.Answer(q, &tb.QueryResponse{
			Results:   nil,
			CacheTime: 60,
		})
		return
	}
	article := &tb.ArticleResult{
		Title:       state.Name,
		Text:        "Updating soon",
		Description: state.Name,
	}

	r := make(tb.Results, 1)
	r[0] = article
	_ = t.bot.Answer(q, &tb.QueryResponse{
		Results:   r,
		CacheTime: 0,
		QueryID:   state.Abbreviation,
	})
}

func (t *Telegram) handleChosenInlineResult(c *tb.ChosenInlineResult) {
	states := election.GetIndexStates()
	_, ok := states[strings.ToUpper(c.Query)]
	if !ok {
		return
	}

}

func (t *Telegram) Stop() {
	t.bot.Stop()
}

func (t *Telegram) runListener() {
	go t.runUpdater()
	for _, s := range election.GetStates() {
		state, err := t.redis.GetMessageIdForState(t.channel.ID, s.Abbreviation)

		if err != nil || state == 0 {
			t.log.Infof("sending new message for %s", s)

			//menu := &tb.ReplyMarkup{ResizeReplyKeyboard: true}
			//
			//menu.InlineKeyboard = [][]tb.InlineButton{
			//	{
			//		{
			//			Text:   "Subscribe?",
			//			Unique: "something",
			//		},
			//	},
			//}

			send, err := t.bot.Send(t.channel, "Hello world, this message is for: "+s.Name)

			if err != nil {
				panic(errors.Wrap(err, "cba to deal with this rn"))
			}

			if err := t.redis.SaveMessageIdForState(t.channel.ID, s.Abbreviation, send.ID); err != nil {
				panic(errors.Wrap(err, "cba to deal with this rn"))
			}

			time.Sleep(time.Second * 4)

			continue
		}
	}

}

func (t *Telegram) runUpdater() {
	m := make(map[string]*StateVote)
	for {
		update := <-t.dataChannel

		for _, vote := range update.Votes {
			val, ok := m[vote.State.Abbreviation]
			if !ok {
				val = &StateVote{}
				m[vote.State.Abbreviation] = val
			}

			if vote.Candidate.Party.Abbreviation == "Dem" {
				val.dem = vote
			}

			if vote.Candidate.Party.Abbreviation == "GOP" {
				val.rep = vote
			}
		}

		for _, val := range m {
			state := val.dem.State.Abbreviation

			id, err := t.redis.GetMessageIdForState(t.channel.ID, state)

			if err != nil || id == 0 {
				continue
			}

			editableMsg := EditableMessage{
				MsgID:     id,
				ChannelID: t.channel.ID,
			}
			_, _ = t.bot.Edit(editableMsg, GetPrettyMessage(val))
		}
	}
}

func GetPrettyMessage(vote *StateVote) string {

	dem := vote.dem
	rep := vote.rep

	return fmt.Sprintf(
		`
%s

%s State Results
%s
%s`,

		getPeekable(vote), dem.State.Name, getCandidateBlock(dem), getCandidateBlock(rep))
}

func getCandidateBlock(vote election.Vote) string {
	return fmt.Sprintf(
		`%s (%s)
	Votes: %d (%.2f%%)
	Electoral Votes: %d
`, vote.Candidate.LastName, vote.Candidate.Party.Symbol, vote.Count, vote.Percentage, vote.ElectoralVotes)
}

func getPeekable(vote *StateVote) string {
	dem := vote.dem
	rep := vote.rep
	return fmt.Sprintf("%s - %s: %d (%.2f%%) %s: %d (%.2f%%)", dem.State.Abbreviation, dem.Candidate.Party.Symbol, dem.Count, dem.Percentage, rep.Candidate.Party.Symbol, rep.Count, rep.Percentage)
}

type StateVote struct {
	dem election.Vote
	rep election.Vote
}

type EditableMessage struct {
	MsgID     int
	ChannelID int64
}

func (e EditableMessage) MessageSig() (messageID string, chatID int64) {
	return strconv.Itoa(e.MsgID), e.ChannelID
}
