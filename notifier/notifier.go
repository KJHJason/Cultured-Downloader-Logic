package notifier

type Notifier interface {
	Alert(msg string)
}
