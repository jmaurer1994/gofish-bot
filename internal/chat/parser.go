package chat

// TODO: not sure if this needs more length/bounds/etc checking
import (
	"fmt"
	"log"
	"strings"
)

type ParsedCommand struct {
	prefix  string
	command string
	args    []string
}

type CommandParser struct {
	commandprefix string
}

type CommandParseError struct {
	message string
}

func (e *CommandParseError) Error() string {
	return fmt.Sprintf("CommandParseError:%s", e.message)
}

func (cmdparser *CommandParser) ParseCommand(s string) (ParsedCommand, error) {
	tokens := strings.Fields(s)
	t := []rune(tokens[0])

	p := string(t[0])
	log.Printf("Prefix: %s\n", p)
	if p != cmdparser.commandprefix {
		return ParsedCommand{}, &CommandParseError{"Prefix mismatch"}
	}

	if !(len(t) > 1) {
		return ParsedCommand{}, &CommandParseError{"No command provided"}
	}

	c := string(t)[1:]

	var a []string

	if len(tokens) > 1 {
		a = tokens[1:]
	}

	return ParsedCommand{
		prefix:  p,
		command: c,
		args:    a,
	}, nil
}
