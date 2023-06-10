package interactive

type Interactive interface {
	StartAction(message string) Action
}
