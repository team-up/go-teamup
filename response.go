package teamup

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Chat struct {
	Team  uint32 `json:"team"`
	User  uint32 `json:"user"`
	Room  uint64 `json:"room"`
	Msg   uint64 `json:"msg"`
	Alert uint32 `json:"alert"`
	Name  string `json:"name"`
}

type Feed struct {
	Team      uint32 `json:"team"`
	User      uint32 `json:"user"`
	FeedGroup uint64 `json:"feedgroup"`
	Feed      uint64 `json:"feed"`
	Reply     uint64 `json:"reply"`
	Inform    uint32 `json:"inform"`
}

type FeedGroup struct {
	Team      uint32 `json:"team"`
	FeedGroup uint64 `json:"feedgroup"`
}

type Inform struct {
	Team      uint32 `json:"team"`
	User      uint32 `json:"user"`
	FeedGroup uint64 `json:"feedgroup"`
	Feed      uint64 `json:"feed"`
	Noti      uint64 `json:"noti"`
	Read      uint32 `json:"read"`
	Watch     uint32 `json:"watch"`
}

type User struct {
	User  uint32   `json:"user"`
	Team  uint32   `json:"team"`
	Users []uint32 `json:"users"`
}

type Event struct {
	Type      string     `json:"type"`
	Chat      *Chat      `json:"chat"`
	Feed      *Feed      `json:"feed"`
	FeedGroup *FeedGroup `json:"feedgroup"`
	Inform    *Inform    `json:"inform"`
	User      *User      `json:"user"`
}

func ReadJSON(res *http.Response, data interface{}) error {
	defer res.Body.Close()

	if code := res.StatusCode; code < http.StatusOK || code >= http.StatusMultipleChoices {
		return fmt.Errorf("http: %v", res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, data)
}
