package main

import (
	"github.com/Adeithe/go-twitch/irc"
	"log"
	"time"
)

func registerIrcHandlers() {
	tic.RegisterHandlers(func(ircReader *irc.Client) {
		ircReader.OnShardReconnect(onShardReconnect)
		ircReader.OnShardServerNotice(onShardServerNotice)
		ircReader.OnShardLatencyUpdate(onShardLatencyUpdate)
		ircReader.OnShardMessage(onChannelMessage)
		ircReader.OnShardRawMessage(onRawMessage)
	})
}

func onShardReconnect(shardID int) {
	log.Printf("Shard #%d reconnected\n", shardID)

	go func() {
		tic.Close()
		log.Printf("Disconnected\n")
		time.Sleep(3 * time.Second)

		if err := tic.Connect(); err != nil {
			log.Printf("Error reconnecting to IRC: %v\n", err)
		}

		registerIrcHandlers()
		log.Printf("Reconnected\n")
	}()

}

func onShardServerNotice(shardID int, sn irc.ServerNotice) {
	log.Printf("Shard #%d recv server notice: %s\n", shardID, sn.Message)
}

func onShardChannelUserNotice(shardID int, n irc.UserNotice) {
	log.Printf("Shard #%d recv user notice: %s\n", shardID, n.Message)
}

func onShardLatencyUpdate(shardID int, latency time.Duration) {
	log.Printf("Shard #%d has %dms ping\n", shardID, latency.Milliseconds())
}

func onChannelMessage(shardID int, msg irc.ChatMessage) {
	if err := cmdproc.ProcessCommand(msg); err != nil {
		log.Printf("Error processing command: %v\n", err)
	}
}

func onRawMessage(shardID int, msg irc.Message) {
	log.Printf("#%s: %s\n", msg.Sender.Username, msg.Raw)
}
