package taskmodel

type SetParamValTask struct {
	Name  string
	Value string
	Type  string
}

type GetParamValTask struct {
	Name []string
}

type GetParamNamesTask struct {
	ParameterPath string
	NextLevel     int
}

type GetParamAttrTask struct {
	Name []string
}

type SetParamAttrTask struct {
	Name               string
	NotificationChange bool
	Notification       int
	AccessListChange   bool
	AccessList         []string
}

type AddTask struct {
	Name string
}

type DeleteTask struct {
	Name string
}

type UploadTask struct {
	CmdKey       string
	FileType     string
	URL          string
	Username     string
	Password     string
	DelaySeconds int
}

type DownloadTask struct {
	CmdKey         string
	FileType       string
	URL            string
	Username       string
	Password       string
	FileSize       int64
	TargetFileName string
	DelaySeconds   int
	SuccessURL     string
	FailureURL     string
}
