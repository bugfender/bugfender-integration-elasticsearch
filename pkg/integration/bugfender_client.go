package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"time"

	"golang.org/x/oauth2"

	"bugfender-integration-elasticsearch/pkg/backoff"
	"bugfender-integration-elasticsearch/pkg/jsonurl"
	"bugfender-integration-elasticsearch/pkg/oauth2util"
)

type Config struct {
	OAuthClientID     string
	OAuthClientSecret string
	ApiUrl            *url.URL
}

type Client struct {
	config      *Config
	configHash  []byte // hash of the configuration
	appID       int64
	tokenSource *oauth2util.TokenSourceSniffer
	httpclient  *http.Client
	nextPageURL url.URL
}

// NewBugfenderClient Creates a Bugfender client to fetch logs from the provided app ID
func NewBugfenderClient(config *Config, appID int64, state []byte) (*Client, error) {
	dm := Client{config: config,
		appID:       appID,
		configHash:  hashConfig(config),
		nextPageURL: makeFirstPageURL(config, appID),
	}

	// if state can be restored, restore it
	var savedState saveState
	var refreshToken string
	if json.Unmarshal(state, &savedState) == nil &&
		bytes.Equal(savedState.ConfigHash, dm.configHash) &&
		savedState.AppID == appID {
		refreshToken = savedState.OAuthRefreshToken
		dm.nextPageURL = url.URL(savedState.NextPageURL)
	}

	// login, reuse token if possible
	tokenSource, err := login(config, refreshToken)
	if err != nil {
		return nil, err
	}
	dm.tokenSource = tokenSource
	dm.httpclient = oauth2.NewClient(context.Background(), dm.tokenSource) // this client provides token auto-refreshes
	return &dm, nil
}

func hashConfig(config *Config) []byte {
	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(config)
	if err != nil {
		log.Fatal("encode error:", err)
	}
	return sha256.New().Sum(b.Bytes())
}

func login(config *Config, refreshToken string) (*oauth2util.TokenSourceSniffer, error) {
	ctx := context.Background()

	authURL := *(config.ApiUrl)
	authURL.Path = path.Join(authURL.Path, "/auth/authorize")
	tokenURL := *(config.ApiUrl)
	tokenURL.Path = path.Join(tokenURL.Path, "/auth/token")

	conf := &oauth2.Config{
		ClientID:     config.OAuthClientID,
		ClientSecret: config.OAuthClientSecret,
		Scopes:       []string{"all"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL.String(),
			TokenURL: tokenURL.String(),
		},
	}

	token := &oauth2.Token{RefreshToken: refreshToken}
	tokenSource := oauth2util.NewTokenSourceSniffer(conf.TokenSource(ctx, token), token)
	_, err := tokenSource.Token() // force first refresh
	if err != nil {
		token, err = oauth2util.AuthCodeTokenFromWeb(ctx, conf)
		if err != nil {
			return nil, err
		}
		tokenSource = oauth2util.NewTokenSourceSniffer(conf.TokenSource(ctx, token), token)
	}
	return tokenSource, nil
}

func makeFirstPageURL(config *Config, appID int64) url.URL {
	// calculate URL for first page
	firstRequestURL := *(config.ApiUrl)
	firstRequestURL.Path = path.Join(firstRequestURL.Path, fmt.Sprintf("/api/app/%d/logs/paginated", appID))
	q := firstRequestURL.Query()
	q.Set("date_range_start", time.Now().Format(time.RFC3339))
	q.Set("page_size", "10000")
	firstRequestURL.RawQuery = q.Encode()
	return firstRequestURL
}

// GetNextPage gets the next page of logs, blocks until there is some data to return
func (dm *Client) GetNextPage(ctx context.Context) ([]Log, error) {
	boff := backoff.NewExponential(5*time.Second, 300*time.Second)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		page, err := dm.getLogsPage(ctx, dm.nextPageURL)
		if err != nil {
			return nil, err
		}
		if page.PreviousURL == nil {
			boff.Wait(ctx)
			continue
		}
		dm.nextPageURL = url.URL(*page.PreviousURL)
		return page.Data, ctx.Err()
	}
}

type saveState struct {
	ConfigHash        []byte
	AppID             int64
	NextPageURL       jsonurl.URL
	OAuthRefreshToken string
}

// GetState returns the client's state so that it can be restored later
func (dm *Client) GetState() []byte {
	state := saveState{
		dm.configHash,
		dm.appID,
		jsonurl.URL(dm.nextPageURL),
		dm.tokenSource.CurrentToken.RefreshToken,
	}
	b, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	return b
}

type page struct {
	Data        []Log        `json:"data"`
	PreviousURL *jsonurl.URL `json:"previous"`
}

func (dm *Client) getLogsPage(ctx context.Context, url url.URL) (*page, error) {
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("preparing request: %s", err)
	}
	req = req.WithContext(ctx)
	resp, err := dm.httpclient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %s", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Println(err)
		}
	}()
	var page page
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %s. response was: %s", resp.Status, string(bodyBytes))
	}
	err = json.Unmarshal(bodyBytes, &page)
	if err != nil {
		return nil, fmt.Errorf("parsing response: %s. response was: %s", err, string(bodyBytes))
	}
	return &page, nil
}
