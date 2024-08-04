package main

import (
	"log"

	"github.com/jmaurer1994/gofish-bot/internal/app"
)

func main() {
	log.Println("Starting....")

	bot := app.NewApp()
	bot.Init()
	bot.Start()
}
