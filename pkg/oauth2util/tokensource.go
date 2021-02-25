package oauth2util

import "golang.org/x/oauth2"

type TokenSourceSniffer struct {
	ts           oauth2.TokenSource
	CurrentToken *oauth2.Token
}

// NewTokenSourceSniffer returns an oauth2.TokenSource that sniffs the last provided token
func NewTokenSourceSniffer(ts oauth2.TokenSource, initialToken *oauth2.Token) *TokenSourceSniffer {
	return &TokenSourceSniffer{ts: ts, CurrentToken: initialToken}
}

func (s *TokenSourceSniffer) Token() (*oauth2.Token, error) {
	t, err := s.ts.Token()
	if err == nil {
		s.CurrentToken = t
	}
	return t, err
}
