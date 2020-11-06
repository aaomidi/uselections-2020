package telegram

import (
	"fmt"
	"github.com/aaomidi/uselections-2020/data"
	"github.com/aaomidi/uselections-2020/election"
	"github.com/aaomidi/uselections-2020/redis"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
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
		Token:  t.token,
		Poller: &tb.LongPoller{Timeout: 15 * time.Second},
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
	t.dataChannel = make(chan data.OutgoingUpdate, 2)

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

	markup := [][]tb.InlineButton{
		{
			{
				Text:        "Share 🔗",
				InlineQuery: state.Abbreviation,
			},
		},
	}

	article.SetReplyMarkup(markup)

	r := make(tb.Results, 1)
	r[0] = article
	_ = t.bot.Answer(q, &tb.QueryResponse{
		Results:   r,
		CacheTime: 0,
		QueryID:   state.Abbreviation,
	})
}

func getShareMarkup(stateAbbreviation string) [][]tb.InlineButton {
	return [][]tb.InlineButton{
		{
			{
				Text:        "Share 🔗",
				InlineQuery: stateAbbreviation,
			},
		},
	}
}

func (t *Telegram) handleChosenInlineResult(c *tb.ChosenInlineResult) {
	states := election.GetIndexStates()
	state, ok := states[strings.ToUpper(c.Query)]
	if !ok {
		return
	}

	_ = t.redis.SaveInlineMessageId(state.Abbreviation, c.MessageID)
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
	lastSent := time.Now().Add(-1 * time.Hour)
	sent := false
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
			currentTime := time.Now()

			if currentTime.Sub(lastSent) < 20*time.Second {
				continue
			}
			sent = true

			state := val.dem.State.Abbreviation

			id, err := t.redis.GetMessageIdForState(t.channel.ID, state)

			if err != nil || id == 0 {
				continue
			}

			editableMsg := EditableMessage{
				MsgID:     strconv.Itoa(id),
				ChannelID: t.channel.ID,
			}

			t.log.Infof("sending update for %s", state)

			//for i := 0; i < 12; i++ {
			//	_, err = t.bot.Edit(editableMsg, GetPrettyMessage(val), tb.ModeHTML)
			//
			//	if err == nil || strings.Contains(err.Error(), "message is not modified") {
			//		break
			//	}
			//
			//	if i < 11 && strings.Contains(err.Error(), "Too Many Requests") {
			//		time.Sleep(time.Second * 5)
			//	} else {
			//		t.log.WithError(err).Warnf("failed updating state %s", state)
			//	}
			//}

			msgs, err := t.redis.GetInlineMessageId(state)

			if err == nil {
				for _, msgId := range msgs {
					editableMsg = EditableMessage{
						MsgID:     msgId,
						ChannelID: 0,
					}
					_, err = t.bot.Edit(editableMsg, GetPrettyMessage(val), tb.ModeHTML, &tb.ReplyMarkup{InlineKeyboard: getShareMarkup(state)})
					t.log.WithError(err).Info("Some error happened")
				}
			}
		}
		if sent {
			lastSent = time.Now()
			sent = false
		}
	}
}

func getPrinter() *message.Printer {
	return message.NewPrinter(language.English)
}
func GetPrettyMessage(vote *StateVote) string {

	dem := vote.dem
	rep := vote.rep

	return fmt.Sprintf(
		`
%s

%s State Results
%s
%s

Last Updated %s
`,

		getPeekable(vote), dem.State.Name, getCandidateBlock(dem), getCandidateBlock(rep), getFormattedTime())
}

func getFormattedTime() string {
	loc, _ := time.LoadLocation("America/New_York")
	str := time.Now().In(loc).Format("15:04 MST")

	return str
}

func getCandidateBlock(vote election.Vote) string {
	return getPrinter().Sprintf(
		`%s <b>%s</b>
	Votes: %d (%.2f%%)
	Electoral Votes: %d
`, vote.Candidate.Party.Symbol, vote.Candidate.LastName, vote.Count, vote.Percentage*100, vote.ElectoralVotes)
}

func getPeekable(vote *StateVote) string {
	dem := vote.dem
	rep := vote.rep
	return getPrinter().Sprintf("%s - %s: %d (%.2f%%) %s: %d (%.2f%%)", dem.State.Abbreviation, dem.Candidate.Party.Symbol, dem.Count, dem.Percentage*100, rep.Candidate.Party.Symbol, rep.Count, rep.Percentage*100)
}

type StateVote struct {
	dem election.Vote
	rep election.Vote
}

type EditableMessage struct {
	MsgID     string
	ChannelID int64
}

func (e EditableMessage) MessageSig() (messageID string, chatID int64) {
	return e.MsgID, e.ChannelID
}
