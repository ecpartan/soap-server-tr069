package soap

import "encoding/xml"

type DeviceId struct {
	Manufacturer string `xml:"Manufacturer"`
	OUI          string `xml:"OUI"`
	ProductClass string `xml:"ProductClass"`
	SerialNumber string `xml:"SerialNumber"`
}

type EnvInfo struct {
	XMLName xml.Name `xml:"SOAP-ENV:Envelope"`
	SOAPENV string   `xml:"xmlns:SOAP-ENV,attr"`
	SOAPENC string   `xml:"xmlns:SOAP-ENC,attr"`
	Xsi     string   `xml:"xmlns:xsi,attr"`
	Xsd     string   `xml:"xmlns:xsd,attr"`
	Cwmp    string   `xml:"xmlns:cwmp,attr"`
}

type HeaderInfo struct {
	Header struct {
		Text string `xml:",chardata"`
		ID   struct {
			Text           string `xml:",chardata"`
			MustUnderstand string `xml:"SOAP-ENV:mustUnderstand,attr"`
		} `xml:"cwmp:ID"`
	} `xml:"SOAP-ENV:Header"`
}

type Header2Info struct {
	Header struct {
		Text string `xml:",chardata"`
		ID   struct {
			Text           string `xml:",chardata"`
			MustUnderstand string `xml:"SOAP-ENV:mustUnderstand,attr"`
		} `xml:"cwmp:ID"`
		NoMoreRequests string `xml:"cwmp:NoMoreRequests"`
	} `xml:"SOAP-ENV:Header"`
}

type InformResponse struct {
	EnvInfo
	HeaderInfo `xml:"Header"`
	Body       struct {
		Text           string `xml:",chardata"`
		InformResponse struct {
			Text         string `xml:",chardata"`
			MaxEnvelopes int    `xml:"MaxEnvelopes"`
		} `xml:"InformResponse"`
	} `xml:"SOAP-ENV:Body"`
}

type GetBody struct {
	Body struct {
		Text               string `xml:",chardata"`
		GetParameterValues struct {
			Text           string `xml:",chardata"`
			ParameterNames struct {
				Text      string   `xml:",chardata"`
				ArrayType string   `xml:"SOAP-ENC:arrayType,attr"`
				String    []string `xml:"string"`
			} `xml:"ParameterNames"`
		} `xml:"cwmp:GetParameterValues"`
	} `xml:"SOAP-ENV:Body"`
}

type setValue struct {
	Text string `xml:",chardata"`
	Type string `xml:"xsi:type,attr"`
}

type setParameterValueStruct struct {
	Text  string   `xml:",chardata"`
	Name  string   `xml:"Name"`
	Value setValue `xml:"Value"`
}

type SetBody struct {
	Body struct {
		Text               string `xml:",chardata"`
		SetParameterValues struct {
			Text          string `xml:",chardata"`
			ParameterList struct {
				Text                    string                    `xml:",chardata"`
				ArrayType               string                    `xml:"SOAP-ENC:arrayType,attr"`
				SetParameterValueStruct []setParameterValueStruct `xml:"ParameterValueStruct"`
			} `xml:"ParameterList"`
			ParameterKey string `xml:"ParameterKey"`
		} `xml:"cwmp:SetParameterValues"`
	} `xml:"SOAP-ENV:Body"`
}

type AddBody struct {
	Body struct {
		Text      string `xml:",chardata"`
		AddObject struct {
			Text         string `xml:",chardata"`
			ObjectName   string `xml:"ObjectName"`
			ParameterKey string `xml:"ParameterKey"`
		} `xml:"cwmp:AddObject"`
	} `xml:"SOAP-ENV:Body"`
}

type DeleteBody struct {
	Body struct {
		Text         string `xml:",chardata"`
		DeleteObject struct {
			Text         string `xml:",chardata"`
			ObjectName   string `xml:"ObjectName"`
			ParameterKey string `xml:"ParameterKey"`
		} `xml:"cwmp:DeleteObject"`
	} `xml:"SOAP-ENV:Body"`
}

type GetRPCMethodsBody struct {
	Body struct {
		Text          string `xml:",chardata"`
		GetRPCMethods struct {
			Text string `xml:",chardata"`
			Cwmp string `xml:"cwmp,attr"`
		} `xml:"cwmp:GetRPCMethods"`
	} `xml:"SOAP-ENV:Body"`
}

type GetParameterNamesBody struct {
	Body struct {
		Text              string `xml:",chardata"`
		GetParameterNames struct {
			Text          string `xml:",chardata"`
			Cwmp          string `xml:"cwmp,attr"`
			ParameterPath string `xml:"ParameterPath"`
			NextLevel     string `xml:"NextLevel"`
		} `xml:"cwmp:GetParameterNames"`
	} `xml:"SOAP-ENV:Body"`
}

type GetParameterAttrBody struct {
	Body struct {
		Text                   string `xml:",chardata"`
		GetParameterAttributes struct {
			Text           string `xml:",chardata"`
			ParameterNames struct {
				Text      string   `xml:",chardata"`
				ArrayType string   `xml:"SOAP-ENC:arrayType,attr"`
				String    []string `xml:"string"`
			} `xml:"ParameterNames"`
		} `xml:"cwmp:GetParameterAttributes"`
	} `xml:"SOAP-ENV:Body"`
}

type SetParameterValues struct {
	EnvInfo
	HeaderInfo
	SetBody
}

type GetParameterValues struct {
	EnvInfo
	HeaderInfo
	GetBody
}

type AddObject struct {
	EnvInfo
	HeaderInfo
	AddBody
}

type DeleteObject struct {
	EnvInfo
	HeaderInfo
	DeleteBody
}

type GetRPCMethods struct {
	EnvInfo
	Header2Info
	GetRPCMethodsBody
}

type GetParameterNames struct {
	EnvInfo
	Header2Info
	GetParameterNamesBody
}

type GetParameterAttributes struct {
	EnvInfo
	Header2Info
	GetParameterAttrBody
}
