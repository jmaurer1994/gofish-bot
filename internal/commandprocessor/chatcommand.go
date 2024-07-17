package commandprocessor

import (
	"time"
)

type Command struct {
	Key          string
	F            CommandFunction
	IsModCommand bool
	cooldown     time.Duration
	onCooldown   bool
}

func (cmd *Command) setOnCooldown(v bool) {
	cmd.onCooldown = v
}

func (cmd *Command) activateCooldown() {
	cmd.setOnCooldown(true)

	go func() {
		time.Sleep(cmd.cooldown)
		cmd.setOnCooldown(false)
	}()
}
