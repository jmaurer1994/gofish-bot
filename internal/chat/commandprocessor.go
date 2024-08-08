package chat

import (
	"fmt"
	"log"

	"github.com/Adeithe/go-twitch/irc"
)

type CommandProcessor struct {
	commands map[string]Command
	parser   CommandParser
}

func NewCommandProcessor(prefix string) *CommandProcessor {
	return &CommandProcessor{
		commands: make(map[string]Command),
		parser:   CommandParser{prefix},
	}
}

type CommandProcessorError struct {
	message string
}

func (e *CommandProcessorError) Error() string {
	return fmt.Sprintf("CommandProcessingError:\t%s", e.message)
}

func (cmdproc *CommandProcessor) ProcessCommand(msg irc.ChatMessage) {
	pc, err := cmdproc.parser.ParseCommand(msg.Text)

	if err != nil {
		return
	}

	cmd, ok := cmdproc.commands[pc.command]

	if !ok {
		log.Printf("[CMD] Command not found [%s][%s]\n", msg.Sender.Username, pc.command)
		return
	}

	if cmd.IsModCommand && !msg.Sender.IsModerator {
		log.Printf("[CMD] Command is moderators only [%s][%s]\n", msg.Sender.Username, pc.command)
		return
	}

	if cmd.onCooldown {
		log.Printf("[CMD] Command on cooldown [%s][%s]\n", msg.Sender.Username, pc.command)
		return
	}
	cmd.activateCooldown()
	log.Printf("[CMD] Executing command [%s][%s]\n", msg.Sender.Username, pc.command)
	go cmd.F(pc.args)

}

func (cmdproc *CommandProcessor) RegisterCommand(cmd Command) {
	cmd.setOnCooldown(false)
	cmdproc.commands[cmd.Key] = cmd
}

func (cmdproc *CommandProcessor) RegisterCommands(cmds ...Command) {
	for _, cmd := range cmds {
		cmdproc.RegisterCommand(cmd)
	}
}

type CommandFunction func(args []string)
