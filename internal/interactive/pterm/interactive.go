package pterm

import (
	"github.com/masinger/incredible/internal/interactive"
	"github.com/pterm/pterm"
)

type ptermInteractive struct {
}

func (p ptermInteractive) Confirm(confirmation interactive.Confirmation) (bool, error) {
	return pterm.DefaultInteractiveConfirm.
		WithDefaultValue(confirmation.Default).
		WithTextStyle(pterm.Warning.MessageStyle).
		Show(confirmation.Message)
}

func (p ptermInteractive) StartAction(message string) interactive.Action {
	pterm.Info.Println(message)
	return &ptermAction{
		message: message,
	}
}

func NewPtermInteractive() interactive.Interactive {
	return &ptermInteractive{}
}
