package command_processor

import (
	"errors"
	"github.com/Adeithe/go-twitch/irc"
	"strings"
	"time"
)

type CommandProcessor struct {
	chatCommands  map[string]ChatCommand
	commandPrefix string
	cooldownMap   map[string]bool
}

func New(prefix string) CommandProcessor {
	return CommandProcessor{
		chatCommands:  make(map[string]ChatCommand),
		commandPrefix: prefix,
	}
}

type ParsedCommand struct {
	prefix string
	key    string
	args   []string
}

func (cmdproc *CommandProcessor) ProcessCommand(msg irc.ChatMessage) error {
	pc, err := cmdproc.parseCommandFromString(msg.Text)

	if err != nil {
		return errors.Join(errors.New("Commad parse error"), err)
	}

	cmd, ok := cmdproc.chatCommands[pc.key]

	if !ok {
		return errors.New("Command not found")
	}

	if cmd.IsModCommand && !msg.Sender.IsModerator {
		return errors.New("Command is moderators only")
	}

	if cmd.onCooldown {
		return errors.New("Command on cooldown")
	}

	cmd.activateCooldown()

	cmd.F(pc.args)

	return nil
}

// TODO: not sure if this needs more length/bounds/etc checking
func (cmdproc *CommandProcessor) parseCommandFromString(s string) (ParsedCommand, error) {
	tokens := strings.Fields(s)

	t := []rune(tokens[0])

	p := string(t[0])

	if p != cmdproc.commandPrefix {
		return ParsedCommand{}, errors.New("Prefix mismatch")
	}

	if !(len(t) > 1) {
		return ParsedCommand{}, errors.New("No command provided")
	}

	k := string(t)[1:]

	var a []string

	if len(tokens) > 1 {
		a = tokens[1:]
	}

	return ParsedCommand{
		prefix: p,
		key:    k,
		args:   a,
	}, nil
}

func (cmdproc *CommandProcessor) RegisterCommand(cmd ChatCommand) {
	cmd.setOnCooldown(false)
	cmdproc.chatCommands[cmd.Key] = cmd
}

func (cmdproc *CommandProcessor) RegisterCommands(cmds ...ChatCommand) {
	for _, cmd := range cmds {
		cmdproc.RegisterCommand(cmd)
	}
}

type CommandFunction func(args []string)

type ChatCommand struct {
	Key          string
	F            CommandFunction
	IsModCommand bool
	Cooldown     time.Duration
	onCooldown   bool
}

func (cc *ChatCommand) setOnCooldown(v bool) {
	cc.onCooldown = v
}

func (cmd *ChatCommand) activateCooldown() {
	cmd.setOnCooldown(true)

	go func() {
		time.Sleep(cmd.Cooldown)
		cmd.setOnCooldown(false)
	}()
}
