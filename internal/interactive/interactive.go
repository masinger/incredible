package interactive

type Interactive interface {
	StartAction(message string) Action
	Confirm(confirmation Confirmation) (bool, error)
}
