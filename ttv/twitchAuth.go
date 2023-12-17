package ttv

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	twitchauth "golang.org/x/oauth2/twitch"
	"log"
	"net/http"
	"os"
	//"time"
	//
)

/*
*

	Will start http server, make request to twitch auth, handle redirect uri, stop server, return token
*/

type TwitchAuthClient struct {
	config             *oauth2.Config
	tokenSourceChannel chan oauth2.TokenSource
	httpPort           string
	server             *http.Server
	tokenKey           string
}

func NewTwitchTokenSource(tokenKey, clientID, clientSecret, redirectUri, port string, requiredScopes []string) oauth2.TokenSource {
	tc := &TwitchAuthClient{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       requiredScopes,
			Endpoint:     twitchauth.Endpoint,
			RedirectURL:  redirectUri,
		},
		tokenSourceChannel: make(chan oauth2.TokenSource, 1),
		tokenKey:           tokenKey,
		httpPort:           port,
	}
	router := mux.NewRouter()

	router.HandleFunc("/", tc.HandleRoot).Methods("GET")
	router.HandleFunc("/login", tc.HandleLogin).Methods("GET")
	router.HandleFunc("/redirect", tc.HandleOAuth2Callback).Methods("GET")

	tc.server = &http.Server{Addr: ":" + tc.httpPort, Handler: router}

	return tc.GetTwitchTokenSource()
}

// saveToken saves an OAuth2 token to a file.
func saveToken(tokenKey string, token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return os.WriteFile(".tokens/"+tokenKey, data, 0600)
}

// loadToken loads an OAuth2 token from a file.
func loadToken(tokenKey string) (*oauth2.Token, error) {
	data, err := os.ReadFile(".tokens/" + tokenKey)
	if err != nil {
		return nil, err
	}
	token := &oauth2.Token{}
	err = json.Unmarshal(data, token)
	return token, err
}

func (tc *TwitchAuthClient) GetTwitchTokenSource() oauth2.TokenSource {
	ctx := context.Background()

	var token *oauth2.Token
	var err error

	// Try to load the token from the file
	token, err = loadToken(tc.tokenKey)
	if err != nil {
		// Handle the case where the token does not exist or is invalid
		// You might need to obtain a new token here

		log.Printf("Login required, scopes: {%s}\nStarted running on http://localhost:%s\n", tc.config.Scopes, tc.httpPort)

		tc.startHttpServer()

		tokenSource := <-tc.tokenSourceChannel

		tc.stopHttpServer()

		token, err = tokenSource.Token()

		if err != nil {
			log.Fatalf("Could not get auth token from source")
		}

		saveToken(tc.tokenKey, token)
		return tokenSource
	}

	// Reuse the token source and refresh the token if necessary
	tokenSource := tc.config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token() // This will refresh the token if it's expired
	if err != nil {
		// Handle erro
		fmt.Printf("%v\n", err)
	}

	// Save the potentially refreshed token back to the file
	if err := saveToken(tc.tokenKey, newToken); err != nil {
		// Handle error
		fmt.Printf("%v\n", err)
	}

	return tokenSource
}

func (tc *TwitchAuthClient) startHttpServer() {
	//use a channel to wait for token
	go func() {
		if err := tc.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	return
}

func (tc *TwitchAuthClient) stopHttpServer() {
	ctx := context.Background()
	tc.server.Shutdown(ctx)
}

func (tc *TwitchAuthClient) HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<html><body><a href="/login">Login using Twitch</a></body></html>`))

	return
}

// HandleLogin is a Handler that redirects the user to Twitch for login, and provides the 'state'
// parameter which protects against login CSRF.
func (tc *TwitchAuthClient) HandleLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, tc.config.AuthCodeURL(""), http.StatusTemporaryRedirect)

	return
}

// HandleOauth2Callback is a Handler for oauth's 'redirect_uri' endpoint;
// it validates the state token and retrieves an OAuth token from the request parameters.
func (tc *TwitchAuthClient) HandleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	token, err := tc.config.Exchange(ctx, r.FormValue("code"))
	if err != nil {
		log.Fatalf("error during token exchange\n")

	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

	tokenSource := tc.config.TokenSource(ctx, token)
	tc.tokenSourceChannel <- tokenSource
	return
}
