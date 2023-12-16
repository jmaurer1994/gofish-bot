package ttv

import (
	"fmt"
	"github.com/Adeithe/go-twitch"
	"github.com/Adeithe/go-twitch/irc"
	"golang.org/x/oauth2"
	"log"
)

type TwitchIrcClient struct {
	Channel     string
	Username    string
	writer      irc.Conn
	reader      *irc.Client
	TokenSource oauth2.TokenSource
	token       oauth2.Token
}

func (tic *TwitchIrcClient) Connect() error {
	//setup IRC reader/writeri

	token, err := tic.TokenSource.Token()

	if err != nil {
		return fmt.Errorf("Error connecting to Twitch IRC\n")
	}

	tic.token = *token

	tic.writer = irc.Conn{}
	tic.writer.SetLogin(tic.Username, "oauth:"+tic.token.AccessToken)

	if err := tic.writer.Connect(); err != nil {
        return fmt.Errorf("Error connecting writer: %v", err)	
	}

	tic.reader = twitch.IRC()

	if err := tic.reader.Join(tic.Channel); err != nil {
        return fmt.Errorf("Error joining as reader: %v", err)
	}
	log.Printf("<Connected to channel>[%s:%s]\n", tic.Channel, tic.Username)
	return nil
}

func (tic *TwitchIrcClient) Close() {
	tic.writer.Close()
	tic.reader.Close()
}

type HandlerRegistrationFunction func(*irc.Client)

func (tic *TwitchIrcClient) RegisterHandlers(f HandlerRegistrationFunction) {
	f(tic.reader)
}

func (tic *TwitchIrcClient) SendChannelMessage(message string) {
	tic.writer.Say(tic.Channel, message)
}
