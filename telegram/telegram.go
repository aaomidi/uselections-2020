package telegram

import (
	tb "gopkg.in/tucnak/telebot.v2"
	"time"
)

type Telegram struct {
	token string
	bot   *tb.Bot
}

func New(token string) Telegram {
	return Telegram{
		token: token,
		bot:   nil,
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

	return nil
}

func (t *Telegram) Start() error {
	// On inline query
	t.bot.Handle(tb.OnQuery, func(q *tb.Query) {

	})

	t.bot.Handle(tb.OnChosenInlineResult, func(c *tb.ChosenInlineResult) {

	})

	t.bot.Start()

	return nil
}

func (t *Telegram) Stop() {
	t.bot.Stop()
}
