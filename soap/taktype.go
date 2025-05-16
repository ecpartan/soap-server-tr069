package soap

type TaskType int

const (
	Inform TaskType = iota
	GetParameterValuesResponse
	SetParameterValuesResponse
	AddObjectResponse
	DeleteObjectResponse
	ResponseUndefinded
)
