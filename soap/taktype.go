package soap

type TaskResponseType int

const (
	Inform TaskResponseType = iota
	GetParameterValuesResponse
	SetParameterValuesResponse
	GetRPCMethodsResponse
	GetParameterNamesResponse
	GetParameterAttributesResponse
	SetParameterAttributesResponse
	FactoryResetResponse
	UploadResponse
	DownloadResponse
	AddObjectResponse
	DeleteObjectResponse
	TransferCompleteResponse
	ResponseUndefinded
	FaultResponse
)
