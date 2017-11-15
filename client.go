package teamup

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"golang.org/x/oauth2"
)

type evConfig struct {
	LpIdleTime    uint32 `json:"lp_idle_time"`
	LpWaitTimeout uint32 `json:"lp_wait_timeout"`
}

type evResult struct {
	Events []Event `json:"events"`
}

type Client struct {
	httpClient *http.Client

	oAuthConfig *oauth2.Config
	oAuthToken  *oauth2.Token

	config   *Config
	evConfig *evConfig
	evURL    string

	mutex  *sync.RWMutex
	chOnce *sync.Once
	ch     chan *Event
}

func (c *Client) Channel() <-chan *Event {
	c.chOnce.Do(func() {
		c.ch = make(chan *Event)
		go c.subscribe()
	})

	return c.ch
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var res *http.Response
	op := func() error {
		var err error
		if res, err = c.getClient().Do(req); err != nil {
			if !c.getToken().Valid() {
				if err := c.passwordCredentials(); err != nil {
					return backoff.Permanent(err)
				}
			}

			return err
		}

		if code := res.StatusCode; code >= http.StatusBadRequest && code < http.StatusInternalServerError {
			return backoff.Permanent(fmt.Errorf("http: %v", res.Status))
		}

		return err
	}

	return res, backoff.Retry(op, backoff.NewExponentialBackOff())
}

func (c *Client) getClient() *http.Client {
	defer c.mutex.RUnlock()
	c.mutex.RLock()

	return c.httpClient
}

func (c *Client) getToken() *oauth2.Token {
	defer c.mutex.RUnlock()
	c.mutex.RLock()

	return c.oAuthToken
}

func (c *Client) subscribe() {
	defer close(c.ch)

	idleTimeout := time.Duration(c.evConfig.LpIdleTime) * time.Second

	for {
		req, err := http.NewRequest(http.MethodGet, c.evURL, nil)
		if err != nil {
			return
		}

		res, err := c.Do(req)
		if err != nil {
			return
		}

		data := evResult{}
		err = ReadJSON(res, &data)
		if err != nil {
			return
		}

		if len(data.Events) <= 0 {
			time.Sleep(idleTimeout)
			continue
		}

		for _, v := range data.Events {
			if v.Type == "user.password" || v.Type == "user.drop" {
				return
			}

			c.ch <- &v
		}
	}
}

func (c *Client) passwordCredentials() error {
	token, err := c.oAuthConfig.PasswordCredentialsToken(context.Background(), c.config.Username, c.config.Password)
	if err != nil {
		return err
	}

	if !token.Valid() {
		return fmt.Errorf("oauth2: %v", token.Extra("error"))
	}

	defer c.mutex.Unlock()
	c.mutex.Lock()

	c.httpClient = c.oAuthConfig.Client(context.Background(), token)
	c.httpClient.Timeout = time.Duration(c.evConfig.LpWaitTimeout+extraLpWaitTimeout) * time.Second

	c.oAuthToken = token

	return nil
}
