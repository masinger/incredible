package interactive

import "fmt"

var NonInteractive = fmt.Errorf("non interactive session")

type Interactive interface {
	StartAction(message string) Action
	Confirm(confirmation Confirmation) (bool, error)
	Input(input Input) (string, error)
}
