package progress

type ProgressBar interface {
	// Add adds the given number to the progress.
	Add(int)

	// Start starts the progress.
	Start()

	// Stop stops the progress.
	// Requires a bool to indicate if there is an error or not after the progress is stopped.
	Stop(bool)

	// Change to spinner, meaning there is no clear indicator of the progress.
	SetToSpinner()

	// Change to progress bar, meaning there is a clear indicator of the progress.
	SetToProgressBar()

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

	// Mainly for frontend usage.
	SnapshotTask() // Saves the current progress state for nested progress bars.
	UpdateFolderPath(string)   // The folder path that the contents will be downloaded to.
}
