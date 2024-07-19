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

func (tic *TwitchIrcClient) refreshAuthToken() error {
	token, err := tic.TokenSource.Token()
	if err != nil {
		return err
	}

	tic.token = *token
	return nil
}

func (tic *TwitchIrcClient) ConnectWriter() error {
	if err := tic.refreshAuthToken(); err != nil {
		return fmt.Errorf("Error retrieving authentication token\n")
	}

	tic.writer = irc.Conn{}
	tic.writer.SetLogin(tic.Username, "oauth:"+tic.token.AccessToken)

	if err := tic.writer.Connect(); err != nil {
		return fmt.Errorf("Error connecting writer: %v", err)
	}

	return nil
}

func (tic *TwitchIrcClient) CloseWriter() {
	tic.writer.Close()
}

func (tic *TwitchIrcClient) ConnectReader() error {
	tic.reader = twitch.IRC()

	if err := tic.ConnectToChannel(); err != nil {
		return fmt.Errorf("Error joining channel as reader: %v\n", err)
	}

	return nil

}

func (tic *TwitchIrcClient) CloseReader() {
	tic.reader.Close()
}

func (tic *TwitchIrcClient) InitializeConnection() error {
	if err := tic.ConnectWriter(); err != nil {
		return fmt.Errorf("Error Connecting Writer: %v", err)
	}

	if err := tic.ConnectReader(); err != nil {
		return fmt.Errorf("Error Connecting Reader %v", err)
	}

	return nil
}

func (tic *TwitchIrcClient) CloseConnection() {
	tic.CloseReader()
	tic.CloseWriter()
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

func (tic *TwitchIrcClient) DisconnectFromChannel() error {
	if err := tic.reader.Leave(tic.Channel); err != nil {
		return err
	}

	return nil
}

func (tic *TwitchIrcClient) ReaderIsConnected() (irc.RoomState, bool) {
	return tic.reader.GetChannel(tic.Channel)
}

func (tic *TwitchIrcClient) Sendf(message string, a ...interface{}) error {
	return tic.writer.Sayf(tic.Channel, message, a...)
}
