package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type asanaTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	UserID       string // asana gid
}

func (s *server) tokensForRequest(r *http.Request) (*asanaTokens, error) {
	sid := getSessionCookie(r)
	if sid == "" {
		return nil, errors.New("no session")
	}
	var t asanaTokens
	err := s.db.QueryRow(`
		select oa.access_token, oa.refresh_token, coalesce(oa.expires_at, now()), s.user_id
		from sessions s
		join oauth_accounts oa on oa.user_id = s.user_id and oa.provider='asana'
		where s.id=$1`, sid).Scan(&t.AccessToken, &t.RefreshToken, &t.ExpiresAt, &t.UserID)
	if err != nil {
		return nil, err
	}
	// refresh a little early
	if time.Now().After(t.ExpiresAt.Add(-2 * time.Minute)) && t.RefreshToken != "" {
		if err := s.refreshAsanaTokens(&t); err != nil {
			return nil, err
		}
	}
	return &t, nil
}

func (s *server) refreshAsanaTokens(t *asanaTokens) error {
	cfg := loadOAuth()
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", cfg.clientID)
	form.Set("client_secret", cfg.clientSecret)
	form.Set("refresh_token", t.RefreshToken)

	req, _ := http.NewRequest("POST", "https://app.asana.com/-/oauth_token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil { return err }
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return errors.New("refresh failed: " + string(b))
	}

	var tok struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int64  `json:"expires_in"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&tok); err != nil { return err }

	t.AccessToken = tok.AccessToken
	if tok.RefreshToken != "" {
		t.RefreshToken = tok.RefreshToken
	}
	t.ExpiresAt = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)

	_, err = s.db.Exec(`update oauth_accounts
		set access_token=$1, refresh_token=$2, expires_at=$3
		where user_id=$4 and provider='asana'`,
		t.AccessToken, t.RefreshToken, t.ExpiresAt, t.UserID)
	return err
}

func (s *server) asanaGET(r *http.Request, path string, q url.Values, out any) error {
	t, err := s.tokensForRequest(r)
	if err != nil { return err }
	if q == nil { q = url.Values{} }
	u := "https://app.asana.com/api/1.0" + path
	if len(q) > 0 { u += "?" + q.Encode() }

	req, _ := http.NewRequestWithContext(context.Background(), "GET", u, nil)
	req.Header.Set("Authorization", "Bearer "+t.AccessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil { return err }
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return errors.New("asana GET failed: " + string(b))
	}
	return json.NewDecoder(res.Body).Decode(out)
}

