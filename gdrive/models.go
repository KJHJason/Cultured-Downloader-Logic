package gdrive

type GDriveToDl struct {
	Id       string
	Type     string
	FilePath string
}

type GdriveFileToDl struct {
	Id          string
	Name        string
	Size        int64
	MimeType    string
	Md5Checksum string
	FilePath    string
}

type GdriveError struct {
	Err      error
	FilePath string
}
