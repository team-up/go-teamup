package teamup

import (
	"net/http"
	"net/url"
	"strings"
	"sync"

	"golang.org/x/oauth2"
)

const (
	DefaultEdgeHost = "edge.tmup.com"
	DefaultEvHost   = "ev.tmup.com"
	DefaultAuthHost = "auth.tmup.com"

	defaultLpIdleTime    = 1  // default long polling idle sleep time (seconds)
	defaultLpWaitTimeout = 30 // default long polling wait timeout (seconds)

	extraLpWaitTimeout = 5
)

type Config struct {
	EdgeHost string
	EvHost   string
	AuthHost string

	ClientID     string
	ClientSecret string

	Username string
	Password string

	EvPrefixFilters []string
}

func NewClient(config *Config) (*Client, error) {
	if len(config.EdgeHost) <= 0 {
		config.EdgeHost = DefaultEdgeHost
	}

	if len(config.EvHost) <= 0 {
		config.EvHost = DefaultEvHost
	}

	if len(config.AuthHost) <= 0 {
		config.AuthHost = DefaultAuthHost
	}

	// get ev configuration
	evURL := "https://" + config.EvHost
	res, err := http.DefaultClient.Get(evURL)
	if err != nil {
		return nil, err
	}

	evConfig := &evConfig{
		LpIdleTime:    defaultLpIdleTime,
		LpWaitTimeout: defaultLpWaitTimeout,
	}
	if err := ReadJSON(res, evConfig); err != nil {
		return nil, err
	}

	evURL += "/v3/events"
	if len(config.EvPrefixFilters) > 0 {
		urlValues := url.Values{}
		urlValues.Set("prefix", strings.Join(config.EvPrefixFilters, "|"))
		evURL += "?" + urlValues.Encode()
	}

	// escape basic auth
	oauth2.RegisterBrokenAuthHeaderProvider("https://" + config.AuthHost + "/oauth2/")

	oAuthConfig := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://" + config.AuthHost + "/oauth2/authorize",
			TokenURL: "https://" + config.AuthHost + "/oauth2/token",
		},
	}

	client := &Client{
		oAuthConfig: oAuthConfig,
		config:      config,
		evConfig:    evConfig,
		evURL:       evURL,
		mutex:       &sync.RWMutex{},
		chOnce:      &sync.Once{},
	}

	if err = client.passwordCredentials(); err != nil {
		return nil, err
	}

	return client, nil
}
