package ttv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"net/http"
)

type TwitchApiClient struct {
	ClientId      string
	BroadcasterId string
	TokenSource   oauth2.TokenSource
	token         oauth2.Token
}

func (tac *TwitchApiClient) RefreshToken() error {
	token, err := tac.TokenSource.Token()

	if err != nil {
		return fmt.Errorf("Error refreshing token: %v", err)
	}

	tac.token = *token

	return nil
}

func (tac *TwitchApiClient) SetChannelTitle(newTitle string) error {

	// Twitch API endpoint for updating channel information
	url := "https://api.twitch.tv/helix/channels"

	// fmt.Printf("<%s><%s>[%s]: %s\n", clientId, broadcasterId, accessToken, newTitle)
	// Set up the request payload
	payload := map[string]string{
		"broadcaster_id": tac.BroadcasterId,
		"title":          newTitle,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshalling payload: %v", err)
	}

	// Create the request
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set the necessary headers
	req.Header.Set("Authorization", "Bearer "+tac.token.AccessToken)
	req.Header.Set("Client-Id", tac.ClientId) // Replace with your client ID
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()
	// Check for a successful status code
	if resp.StatusCode != http.StatusNoContent {
        if(resp.StatusCode == http.StatusUnauthorized){
           tac.RefreshToken()
           return tac.SetChannelTitle(newTitle)
        }

		return fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}
	return nil
}
