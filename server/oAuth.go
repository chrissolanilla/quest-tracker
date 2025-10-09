package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
	"io"
)

type oauthConfig struct {
	clientID	 string
	clientSecret string
	redirectURI  string
	scopes	   string
}

func loadOAuth() oauthConfig {
	return oauthConfig{
		clientID:	 getenv("ASANA_CLIENT_ID", ""),
		clientSecret: getenv("ASANA_CLIENT_SECRET", ""),
		redirectURI:  getenv("ASANA_REDIRECT_URI", ""),
		scopes:	   getenv("ASANA_SCOPES", "users:read"),
	}
}

func randomString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func setSessionCookie(w http.ResponseWriter, sid string) {
	http.SetCookie(w, &http.Cookie{
		Name:	 "sid",
		Value:	sid,
		Path:	 "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		//enable when https
		// Secure: true,
	})
}

func getSessionCookie(r *http.Request) string {
	c, _ := r.Cookie("sid")
	if c == nil { return "" }
	return c.Value
}

func pkce() (codeVerifier, codeChallenge string) {
	ver := randomString(32)
	sum := sha256.Sum256([]byte(ver))
	chal := base64.RawURLEncoding.EncodeToString(sum[:])
	return ver, chal
}


func (s *server) mountAuth(mux *http.ServeMux, origin string) {
	mux.HandleFunc("GET /auth/asana/start", s.handleAsanaStart)
	mux.HandleFunc("GET /auth/asana/callback", s.handleAsanaCallback)
}

func (s *server) handleAsanaStart(w http.ResponseWriter, r *http.Request) {
	cfg := loadOAuth()
	if cfg.clientID == "" || cfg.redirectURI == "" {
		http.Error(w, "oauth not configured", 500); return
	}

	state := randomString(24)
	codeVerifier, codeChallenge := pkce()
	sid := getSessionCookie(r)
	if sid == "" { sid = randomString(24) }

	//store these transient values server-side keyed by session id (we'll keep it in memory for dev)
	//for simplicity here: stash them in a temp table? use a simple map? we'll use table "sessions_meta"
	//to avoid global maps. create table if it doesn't exist:
	_, _ = s.db.Exec(`create table if not exists sessions_meta (id text primary key, state text, code_verifier text, created_at timestamptz default now())`)
	_, _ = s.db.Exec(`insert into sessions_meta(id, state, code_verifier) values($1,$2,$3)
		on conflict (id) do update set state=excluded.state, code_verifier=excluded.code_verifier`, sid, state, codeVerifier)

	setSessionCookie(w, sid)

	v := url.Values{}
	v.Set("client_id", cfg.clientID)
	v.Set("redirect_uri", cfg.redirectURI)
	v.Set("response_type", "code")
	v.Set("state", state)
	v.Set("code_challenge_method", "S256")
	v.Set("code_challenge", codeChallenge)
	if cfg.scopes != "" {
		v.Set("scope", cfg.scopes)
	}

	authURL := "https://app.asana.com/-/oauth_authorize?" + v.Encode()
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (s *server) handleAsanaCallback(w http.ResponseWriter, r *http.Request) {
	cfg := loadOAuth()
	q := r.URL.Query()
	code := q.Get("code")
	state := q.Get("state")
	if code == "" || state == "" {
		http.Error(w, "missing code/state", 400); return
	}

	sid := getSessionCookie(r)
	if sid == "" {
		http.Error(w, "no session", 400); return
	}

	var wantState, codeVerifier string
	err := s.db.QueryRow(`select state, code_verifier from sessions_meta where id=$1`, sid).Scan(&wantState, &codeVerifier)
	if err != nil || wantState != state {
		http.Error(w, "state mismatch", 400); return
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", cfg.clientID)
	form.Set("client_secret", cfg.clientSecret)
	form.Set("redirect_uri", cfg.redirectURI)
	form.Set("code", code)
	form.Set("code_verifier", codeVerifier)

	req, _ := http.NewRequest("POST", "https://app.asana.com/-/oauth_token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil { http.Error(w, err.Error(), 502); return }
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		// http.Error(w, res.Status, 502); return
		b, _ := io.ReadAll(res.Body)
		http.Error(w, "bad token response: "+string(b), 502)
		return
	}

	var tok struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn	int64  `json:"expires_in"`
		TokenType	string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		Data struct {
			Gid   string `json:"gid"`
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&tok); err != nil {
		http.Error(w, "bad token response", 502); return
	}
	if tok.AccessToken == "" {
		http.Error(w, "no access token", 502); return
	}

	user := tok.Data
	if user.Gid == "" {
		//fall back to /users/me if not present
		req2, _ := http.NewRequestWithContext(context.Background(), "GET", "https://app.asana.com/api/1.0/users/me", nil)
		req2.Header.Set("Authorization", "Bearer "+tok.AccessToken)
		res2, err := http.DefaultClient.Do(req2)
		if err == nil && res2.StatusCode == 200 {
			defer res2.Body.Close()
			var wrap struct{ Data struct {
				Gid string `json:"gid"`
				Name string `json:"name"`
				Email string `json:"email"`
			} `json:"data"` }
			_ = json.NewDecoder(res2.Body).Decode(&wrap)
			user = wrap.Data
		}
	}

	//upsert user and oauth account
	_, _ = s.db.Exec(`insert into users(id, name, avatar_url)
					  values($1,$2,$3)
					  on conflict (id) do update set name=excluded.name`,
		user.Gid, user.Name, "")

	expiresAt := time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
	_, _ = s.db.Exec(`
		insert into oauth_accounts(user_id, provider, access_token, refresh_token, scope, expires_at)
		values($1,'asana',$2,$3,$4,$5)
		on conflict (user_id, provider) do update set
		  access_token=excluded.access_token,
		  refresh_token=excluded.refresh_token,
		  scope=excluded.scope,
		  expires_at=excluded.expires_at`,
		user.Gid, tok.AccessToken, tok.RefreshToken, getenv("ASANA_SCOPES", ""), expiresAt)

	_, _ = s.db.Exec(`insert into sessions(id, user_id) values($1,$2)
		on conflict (id) do update set user_id=excluded.user_id`, sid, user.Gid)

	_, _ = s.db.Exec(`delete from sessions_meta where id=$1`, sid)

	go func(r0 *http.Request, uid string) {
		_ = s.recomputePointsForUser(r0, uid)
	}(r.Clone(r.Context()), user.Gid)


	http.Redirect(w, r, getenv("POST_LOGIN_REDIRECT", "http://localhost:5173/profile"), http.StatusFound)
}

