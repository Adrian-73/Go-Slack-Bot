package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func main() {
	er := godotenv.Load(".env")
	if er != nil {
		log.Fatal("Error loading .env file")
	}

	BotToken := os.Getenv("BotToken")
	AppToken := os.Getenv("AppToken")

	Client := slack.New(BotToken, slack.OptionDebug(true),
		slack.OptionAppLevelToken(AppToken))

	fmt.Println(Client.AuthTest())
	Socket := socketmode.New(Client,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode : ", log.Lshortfile|log.LstdFlags)),
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func(ctx context.Context, Client *slack.Client, Socket *socketmode.Client) {

		for {
			select {

			case <-ctx.Done():
				log.Println("stopping SocketMode event handler")
				return
			case evt := <-Socket.Events:

				switch evt.Type {

				case socketmode.EventTypeEventsAPI:

					eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
					if !ok {
						fmt.Printf("Ignored %+v", evt)
						continue
					}

					Socket.Ack(*evt.Request)
					err := HandleEventMessage(eventsAPIEvent, Client)

					if err != nil {
						log.Fatal(err)
						continue
					}
					HandleAppMentionEventToBot(eventsAPIEvent.InnerEvent.Data.(*slackevents.AppMentionEvent), Client)

				}
			}
		}
	}(ctx, Client, Socket)
	Socket.Run()

}

func HandleEventMessage(event slackevents.EventsAPIEvent, client *slack.Client) error {
	switch event.Type {
	case slackevents.CallbackEvent:

		innerEvent := event.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			err := HandleAppMentionEventToBot(ev, client)
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("unsupported event type")
	}
	return nil
}

func HandleAppMentionEventToBot(event *slackevents.AppMentionEvent, client *slack.Client) error {

	user, err := client.GetUserInfo(event.User)
	if err != nil {
		return err
	}

	text := strings.ToLower(event.Text)

	attachment := slack.Attachment{}

	if strings.Contains(text, "hello") || strings.Contains(text, "hi") {

		attachment.Text = fmt.Sprintf("Hello %s", user.Name)
		attachment.Color = "#4af030"
	} else if strings.Contains(text, "download yt") {
		ls := strings.Split(text, " ")
		for i := range ls {
			if strings.Contains(ls[i], "https://www.youtube.com/watch?v=") {
				attachment.Text = fmt.Sprintf("Download link -> %s", ls[i])
				attachment.Color = "#444444 " //red
				break
			}
		}
	}
	_, _, err = client.PostMessage(event.Channel, slack.MsgOptionAttachments(attachment))
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}
	return nil
}
