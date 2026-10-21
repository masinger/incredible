package pterm

import "github.com/pterm/pterm"

type ptermAction struct {
	message string
}

func (p ptermAction) Complete(message string) {
	if message != "" {
		pterm.Success.Println(message)
	} else {
		pterm.Success.Printfln("Done: %s", p.message)
	}
}
