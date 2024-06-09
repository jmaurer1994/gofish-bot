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
    log.Printf("Token expiry: %v", token.Expiry)
	tic.token = *token
    
	tic.writer = irc.Conn{}
	tic.writer.SetLogin(tic.Username, "oauth:"+tic.token.AccessToken)

	if err := tic.writer.Connect(); err != nil {
        return fmt.Errorf("Error connecting writer: %v", err)	
	}

	tic.reader = twitch.IRC()

    
	if err := tic.ConnectToChannel(); err != nil {
        return fmt.Errorf("Error joining channel as reader: %v\n", err)
	}

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

func (tic *TwitchIrcClient) ConnectToChannel() error {
   if err := tic.reader.Join(tic.Channel); err != nil {
        return err
	}
	log.Printf("<Connected to channel>[%s:%s]\n", tic.Channel, tic.Username)

    return nil
}

func (tic *TwitchIrcClient) ReaderIsConnected() (irc.RoomState, bool){
    return tic.reader.GetChannel(tic.Channel)
}
