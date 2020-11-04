package telegram

import (
	"github.com/aaomidi/uselections-2020/data"
	"github.com/aaomidi/uselections-2020/redis"
	log "github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
	"time"
)

type Telegram struct {
	token     string
	bot       *tb.Bot
	channelID string
	channel   *tb.Chat
	log       *log.Entry
	redis     *redis.Redis
	data      *data.Data
}

func New(token string, channelId string, r *redis.Redis, d *data.Data) Telegram {
	return Telegram{
		token:     token,
		bot:       nil,
		channelID: channelId,
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
	// On inline query
	t.bot.Handle(tb.OnQuery, t.handleQuery)

	// On inline chosen result
	t.bot.Handle(tb.OnChosenInlineResult, t.handleChosenInlineResult)

	// Start the bot, listen for queries
	t.bot.Start()
}

func (t *Telegram) handleQuery(q *tb.Query) {

}

func (t *Telegram) handleChosenInlineResult(c *tb.ChosenInlineResult) {

}

func (t *Telegram) Stop() {
	t.bot.Stop()
}
