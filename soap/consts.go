package soap

const (
	SoapVersion11 = "1.1"
	SoapVersion12 = "1.2"

	SoapContentType11 = "text/xml; charset=\"utf-8\""
	SoapContentType12 = "application/soap+xml; charset=\"utf-8\""

	NamespaceSoap11 = "http://schemas.xmlsoap.org/soap/envelope/"
	NamespaceSoap12 = "http://www.w3.org/2003/05/soap-envelope"
)

var (
	BNamespaceSoap11 = []byte("http://schemas.xmlsoap.org/soap/envelope/")
	BNamespaceSoap12 = []byte("http://www.w3.org/2003/05/soap-envelope")

	BNamespaceEnc  = []byte("http://schemas.xmlsoap.org/soap/encoding/")
	BNamespaceXsi  = []byte("http://www.w3.org/2001/XMLSchema-instance")
	BNamespaceXsd  = []byte("http://www.w3.org/2001/XMLSchema")
	BNamespaceCwmp = []byte("urn:dslforum-org:cwmp-1-0")
)

const (
	Pref98    = "InternetGatewayDevice"
	Prev181   = "Device"
	CR_URL    = "InternetGatewayDevice.ManagementServer.ConnectionRequestURL"
	CR_U      = "InternetGatewayDevice.ManagementServer.ConnectionRequestUsername"
	CR_P      = "InternetGatewayDevice.ManagementServer.ConnectionRequestPassword"
	A_AUTHU   = "InternetGatewayDevice.ManagementServer.Username"
	A_AUTHP   = "InternetGatewayDevice.ManagementServer.Password"
	SW_V      = "InternetGatewayDevice.DeviceInfo.SoftwareVersion"
	HW_V      = "InternetGatewayDevice.DeviceInfo.HardwareVersion"
	MODELNAME = "InternetGatewayDevice.DeviceInfo.ModelName"
)
