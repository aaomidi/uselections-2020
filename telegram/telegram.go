package telegram

import (
	"github.com/aaomidi/uselections-2020/data"
	"github.com/aaomidi/uselections-2020/election"
	"github.com/aaomidi/uselections-2020/redis"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
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
	for _, s := range election.GetStates() {
		state, err := t.redis.GetMessageIdForState(t.channel.ID, s.Abbreviation)

		if err != nil || state == 0 {
			t.log.Infof("sending new message for %s", s)

			menu := &tb.ReplyMarkup{ResizeReplyKeyboard: true}

			menu.InlineKeyboard = [][]tb.InlineButton{
				{
					{
						Text:   "Subscribe?",
						Unique: "something",
					},
				},
			}

			send, err := t.bot.Send(t.channel, "Hello world, this message is for: "+s.Name, menu)

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

	for {
		update := <-t.dataChannel

		spew.Dump(update)

		// TODO Messages and UX
	}
}
