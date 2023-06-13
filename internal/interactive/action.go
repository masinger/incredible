package interactive

type Action interface {
	Complete(message string)
}
