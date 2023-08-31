package progress

type Progress interface {
	// Add adds the given number to the progress.
	Add(int)

	// Start starts the progress.
	Start()

	// Stop stops the progress.
	// Requires a bool to indicate if there is an error or not after the progress is stopped.
	Stop(bool)

	// Stops the progress due to an interrupt signal.
	StopInterrupt(string)

	// UpdateBaseMsg updates the base message of the progress.
	UpdateBaseMsg(string)

	// UpdateMax updates the maximum value of the progress.
	UpdateMax(int)

	// Increment increments the progress by 1.
	Increment()

	// Updates the success message of the progress.
	UpdateSuccessMsg(string)

	// Updates the message of the progress in the event of an error.
	UpdateErrorMsg(string)
}
