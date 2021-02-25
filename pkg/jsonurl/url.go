package jsonurl

import (
	"encoding/json"
	"net/url"
)

// Credits: https://play.golang.org/p/3bfC8ao33Z

// URL Is an url.URL that can be JSON-unmarshalled
type URL url.URL

func (j *URL) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	*j = URL(*u)
	return nil
}

func (j URL) MarshalJSON() ([]byte, error) {
	u := url.URL(j)
	return json.Marshal(u.String())
}
