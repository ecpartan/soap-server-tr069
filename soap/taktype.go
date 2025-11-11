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
	RebootResponse
	FactoryResetResponse
	UploadResponse
	DownloadResponse
	AddObjectResponse
	DeleteObjectResponse
	TransferCompleteResponse
	ResponseUndefinded
	FaultResponse
)
