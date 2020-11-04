package cmd

import (
	"github.com/aaomidi/uselections-2020/data"
	scraper2 "github.com/aaomidi/uselections-2020/scraper"
	"github.com/aaomidi/uselections-2020/telegram"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the bot",
	RunE: func(cmd *cobra.Command, args []string) error {
		scraper := scraper2.NPRScraper{}
		broadcaster := data.Data{}
		broadcaster.Start(&scraper)

		tg := telegram.New(viper.GetString("token"), viper.GetString("channel"))

		if err := tg.Create(); err != nil {
			return errors.Wrap(err, "error creating telegram bot")
		}

		go tg.Start()

		terminate := make(chan os.Signal, 1)
		signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)

		toTerminate := <-terminate
		for {
			if toTerminate != nil {
				tg.Stop()
				break
			}
		}

		return nil
	},
}
