package app

import (
	"github.com/Adeithe/go-twitch/irc"
	"log"
	"time"
)

func (app *Config) registerIrcHandlers() {
	app.TwitchIrc.RegisterHandlers(func(ircReader *irc.Client) {
		ircReader.OnShardReconnect(app.onShardReconnect)
		ircReader.OnShardServerNotice(app.onShardServerNotice)
		ircReader.OnShardLatencyUpdate(app.onShardLatencyUpdate)
		ircReader.OnShardMessage(app.onChannelMessage)
		ircReader.OnShardRawMessage(app.onRawMessage)
	})
}

func (app *Config) onShardReconnect(shardID int) {
	log.Printf("[IRC] Shard #%d received reconnect\n", shardID)

	go func() {
		app.TwitchIrc.CloseConnection()
		log.Printf("[IRC] Disconnected\n")
		time.Sleep(3 * time.Second)

		if err := app.TwitchIrc.InitializeConnection(); err != nil {
			log.Printf("[IRC] Error reconnecting: %v\n", err)
		}

		app.registerIrcHandlers()
		log.Printf("[IRC] Reconnected\n")
	}()

}

func (app *Config) onShardServerNotice(shardID int, sn irc.ServerNotice) {
	log.Printf("[IRC] Shard #%d recv server notice: %s\n", shardID, sn.Message)
}

func (app *Config) onShardChannelUserNotice(shardID int, n irc.UserNotice) {
	log.Printf("[IRC] Shard #%d recv user notice: %s\n", shardID, n.Message)
}
func (app *Config) onShareChannelRoomState(shardID int, n irc.RoomState) {
	log.Printf("[IRC] Shard #%d recv room state: [%d]%s\n", shardID, n.ID, n.Name)
}
func (app *Config) onShardLatencyUpdate(shardID int, latency time.Duration) {
	log.Printf("[IRC] Shard #%d has %dms ping\n", shardID, latency.Milliseconds())
}

func (app *Config) onChannelMessage(shardID int, msg irc.ChatMessage) {
	app.CmdProc.ProcessCommand(msg)
}

func (app *Config) onRawMessage(shardID int, msg irc.Message) {
	//log.Printf("#%s: %s\n", msg.Sender.Username, msg.Raw)
}
