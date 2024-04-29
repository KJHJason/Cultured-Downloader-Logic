package notify

type Notifier interface {
	Alert(msg string)
}
