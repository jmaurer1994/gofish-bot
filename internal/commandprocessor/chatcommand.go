package commandprocessor

import (
	"time"
)

type Command struct {
	Key          string
	F            CommandFunction
	IsModCommand bool
	Cooldown     time.Duration
	onCooldown   bool
}

func (cmd *Command) setOnCooldown(v bool) {
	cmd.onCooldown = v
}

func (cmd *Command) activateCooldown() {
	cmd.setOnCooldown(true)

	go func() {
		time.Sleep(cmd.Cooldown)
		cmd.setOnCooldown(false)
	}()
}
