package interpreter

import (
	"io"

	"github.com/chzyer/readline"
)

func (i *Interpreter) REPL() error {
	var line string
	i.replMode = true

	rl, err := readline.New("> ")
	if err != nil {
		i.log.With("error", err).Error("Failed to initialize readline library")
		return err
	}
	defer rl.Close()
	rl.CaptureExitSignal()

	for {
		line, err = rl.Readline()
		if err != nil {
			switch err {
			case readline.ErrInterrupt:
				continue
			case io.EOF:
				i.log.Info("EOF detected, exiting...")
				return nil
			}
			i.log.With("error", err).Error("Unexpected error occurred, exiting")
			return err
		}

		switch line {
		case "exit":
			i.log.Info("Exiting glocks repl")
			return nil
		default: // REPL process line
			if err = i.run(line); err != nil {
				i.log.Warn(err)
			}
		}
	}
}
