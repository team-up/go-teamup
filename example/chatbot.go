package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/team-up/go-teamup"
)

type Team struct {
	Index uint32   `json:"index"`
	Role  []string `json:"role"`
}

type MyInfo struct {
	Index uint32 `json:"index"`
	Teams []Team `json:"teams"`
}

type Button struct {
	Text string `json:"text"`
	Type string `json:"type"`
	URL  string `json:"url,omitempty"`
}

type Extra struct {
	Version      uint32   `json:"version"`
	Type         string   `json:"type"`
	MsgButtons   []Button `json:"msg_buttons,omitempty"`
	InputButtons []Button `json:"input_buttons,omitempty"`
}

type Message struct {
	Content string  `json:"content"`
	Type    uint32  `json:"type,omitempty"`
	Team    uint32  `json:"team,omitempty"`
	User    uint32  `json:"user,omitempty"`
	Extras  []Extra `json:"extras,omitempty"`
}

type ChatBot struct {
	teamupClient *teamup.Client

	botTeams map[uint32]bool
}

// if received "hello" then response "world"
func (c *ChatBot) doChatMessage(chat *teamup.Chat) {
	receivedMessage, err := c.getMessage(chat.Room, chat.Msg)
	if err != nil {
		panic(err)
	}

	fmt.Println(receivedMessage)

	if receivedMessage.Type != 1 {
		return
	}

	message := &Message{}

	if bot, ok := c.botTeams[receivedMessage.Team]; ok && bot {
		// 봇으로 등록된 팀에서만 extras 동작
		message.Content = "click button"
		message.Extras = []Extra{
			{
				Version: 1,
				Type:    "answer",
				InputButtons: []Button{
					{
						Text: "hello",
						Type: "text",
					},
					{
						Text: "do nothing",
						Type: "text",
					},
				},
			},
		}

		if len(receivedMessage.Extras) > 0 {
			for _, extra := range receivedMessage.Extras {
				if extra.Type == "init" {
					info, err := c.postMessage(chat.Room, message)
					if err != nil {
						panic(err)
					}

					fmt.Println(info)
					return
				}
			}
		}
	}

	if strings.EqualFold("hello", receivedMessage.Content) {
		message.Content = "world"

		info, err := c.postMessage(chat.Room, message)
		if err != nil {
			panic(err)
		}

		fmt.Println(info)
	}
}

func (c *ChatBot) getMessage(room uint64, msg uint64) (*Message, error) {
	// http://team-up.github.io/v3/edge/chat/#api-message-getMessageSummary
	getURL := fmt.Sprintf("https://%s/v3/message/summary/%v/%v/1", edgeHost, room, msg)

	req, err := http.NewRequest(http.MethodGet, getURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("all", "1")
	req.URL.RawQuery = q.Encode()

	res, err := c.teamupClient.Do(req)
	if err != nil {
		return nil, err
	}

	message := &Message{}
	err = teamup.ReadJSON(res, &message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (c *ChatBot) postMessage(room uint64, data interface{}) (map[string]interface{}, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// http://team-up.github.io/v3/edge/chat/#api-message-postMessage
	postURL := fmt.Sprintf("https://%s/v3/message/%v", edgeHost, room)

	req, err := http.NewRequest(http.MethodPost, postURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := c.teamupClient.Do(req)
	if err != nil {
		return nil, err
	}

	info := make(map[string]interface{})
	err = teamup.ReadJSON(res, &info)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func NewChatBot(c *teamup.Client) (*ChatBot, error) {
	// http://team-up.github.io/v1/auth/#api-my-getUser
	getURL := fmt.Sprintf("https://%s/v1/user", authHost)

	req, err := http.NewRequest(http.MethodGet, getURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	myInfo := &MyInfo{}
	err = teamup.ReadJSON(res, myInfo)
	if err != nil {
		return nil, err
	}

	chatBot := &ChatBot{
		teamupClient: c,
		botTeams:     make(map[uint32]bool),
	}

	// 봇으로 등록된 팀 목록 설정
	for _, team := range myInfo.Teams {
		for _, role := range team.Role {
			if role == "bot" {
				chatBot.botTeams[team.Index] = true
				break
			}
		}
	}

	return chatBot, nil
}
