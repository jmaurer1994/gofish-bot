package commandprocessor

import (
	"errors"
	"fmt"
	"github.com/Adeithe/go-twitch/irc"
	"log"
)

type CommandProcessor struct {
	commands      map[string]Command
	commandPrefix string
	parser        CommandParser
}

func NewCommandProcessor(prefix string) CommandProcessor {
	return CommandProcessor{
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

func (cmdproc *CommandProcessor) ProcessCommand(msg irc.ChatMessage) error {
	pc, err := cmdproc.parser.ParseCommand(msg.Text)

	if err != nil {
		return nil
	}

	cmd, ok := cmdproc.commands[pc.command]

	if !ok {
		return errors.New("Command not found")
	}

	if cmd.IsModCommand && !msg.Sender.IsModerator {
		return errors.New("Command is moderators only")
	}

	if cmd.onCooldown {
		return errors.New("Command on cooldown")
	}
	log.Println("Activating cooldown")
	cmd.activateCooldown()
	log.Println("Running command function")
	go cmd.F(pc.args)

	return nil
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
