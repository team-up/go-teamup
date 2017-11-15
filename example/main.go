package main

import (
	"fmt"

	"github.com/team-up/go-teamup"
)

const (
	edgeHost = teamup.DefaultEdgeHost
	evHost   = teamup.DefaultEvHost
	authHost = teamup.DefaultAuthHost

	clientID     = "YOUR_CLIENT_ID"
	clientSecret = "YOUR_CLIENT_SECRET"

	username = "YOUR_ACCOUNT"
	password = "YOUR_PASSWORD"
)

func main() {
	c, err := teamup.NewClient(&teamup.Config{
		EvHost:          evHost,
		AuthHost:        authHost,
		ClientID:        clientID,
		ClientSecret:    clientSecret,
		Username:        username,
		Password:        password,
	})
	if err != nil {
		panic(err)
	}

	chatBot, err := NewChatBot(c)
	if err != nil {
		panic(err)
	}

	ch := c.Channel()

	fmt.Println("started")

	for {
		select {
		case ev, ok := <-ch:
			if ok {
				fmt.Println(ev)

				switch ev.Type {
				case "chat.message":
					go chatBot.doChatMessage(ev.Chat)
				}
			} else {
				fmt.Println("client channel closed")
				return
			}
		}
	}

	fmt.Println("exit")
}
