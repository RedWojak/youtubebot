package telegrambot

import (
	"fmt"
	"log"
	"os"
	"time"
	"youtubedownloader/videodownloader"

	tele "gopkg.in/telebot.v4"
)

type Telegrambot struct {
	bot *tele.Bot
}



func (t *Telegrambot) NewBot() (error) {
	pref := tele.Settings{
		Token:  os.Getenv("YTOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}
	
	fmt.Println(pref.Token, ">>>>>>>")
	var err error
	t.bot, err = tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return err
	}

	
	t.bot.Handle("/hello", func(c tele.Context) error {
		return c.Send("Hello!")
	})

	t.bot.Handle(tele.OnText, func(c tele.Context) error {
		videoID, err := videodownloader.Download(c.Update().Message.Text)
		
		fmt.Println("RECEIVED REQUEST: ", c.Update().Message.Text, "\n", "USER ID: ", c.Sender().ID, " USER NAME: ", c.Sender().Username, "\n")


		
		if err != nil {
			return c.Send("Bad request :" + c.Update().Message.Text + "\n" + "Send me valid youtube link")
			return c.Send("Bad request :", c.Update().Message.Text, "\n", "Follow the pattern: https://www.youtube.com/watch?v=pWSMRTLHy1E")
		}

		if videoID == "" {
			return c.Send("Download failed, reason unknown.")
		}

		if videoID == videodownloader.TooLong {
			return c.Send("Download failed, Video is too long, please stick to videos that are no longer then "+time.Duration(videodownloader.MaxDuration).String())
		}

		c.Send("Downloading is done. Sending Video to you NOW!")
		file := &tele.Video{File : tele.FromDisk(videoID)}
		
		return c.Send(file)


	})


	t.bot.Start()
	return nil

}

