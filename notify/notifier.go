package notify

type Notifier interface {
	Alert(msg string)

	Release() // Release the notifier resources
}
