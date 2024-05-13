package progress

// DummyProgBar is a dummy implementation of the ProgressBar interface for testing purposes
type DummyProgBar struct { }

func (tpb *DummyProgBar) Add(int) { }

func (tpb *DummyProgBar) Start() { }

func (tpb *DummyProgBar) Stop(bool) { }

func (tpb *DummyProgBar) SetToSpinner() { }

func (tpb *DummyProgBar) GetIsSpinner() bool { return false }

func (tpb *DummyProgBar) SetToProgressBar() { }

func (tpb *DummyProgBar) GetIsProgBar() bool { return false }

func (tpb *DummyProgBar) StopInterrupt(string) { }

func (tpb *DummyProgBar) UpdateBaseMsg(string) { }

func (tpb *DummyProgBar) UpdateMax(int) { }

func (tpb *DummyProgBar) Increment() { }

func (tpb *DummyProgBar) UpdateSuccessMsg(string) { }

func (tpb *DummyProgBar) UpdateErrorMsg(string) { }

func (tpb *DummyProgBar) SnapshotTask() { }

func (tpb *DummyProgBar) MakeLatestSnapshotMain() { }
