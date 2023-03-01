package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

func main() {
	er := godotenv.Load(".env")
	if er != nil {
		log.Fatalf("Some error occured. Err: %s", er)
	}
	api := slack.New(os.Getenv("SlackToken"))
	_, _, err := api.PostMessage("nothing",
		slack.MsgOptionText("Hello World!", false))
	if err != nil {
		fmt.Printf("%s", err)
	} else {
		fmt.Println("Message sent")
	}

}
