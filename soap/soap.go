package soap

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"
)

// SOAP 1.1 and SOAP 1.2 must expect different ContentTypes and Namespaces.

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

// Envelope type `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
/*type Envelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ SOAP-ENV:Envelope"`
	Header  Header
	Body    SoapBody
}
*/
type Envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Text    string   `xml:",chardata"`
	SOAPENV string   `xml:"SOAP-ENV,attr"`
	SOAPENC string   `xml:"SOAP-ENC,attr"`
	Xsi     string   `xml:"xsi,attr"`
	Xsd     string   `xml:"xsd,attr"`
	Cwmp    string   `xml:"cwmp,attr"`
	Header  struct {
		Text string `xml:",chardata"`
		ID   struct {
			Text           string `xml:",chardata"`
			MustUnderstand string `xml:"mustUnderstand,attr"`
		} `xml:"ID"`
	} `xml:"Header"`
	Body SoapBody `xml:"Body"`
}

// Header type
type Header struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Header"`
	Header  any
}

// Body type
type Body struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`

	Fault               *Fault `xml:",omitempty"`
	Content             any    `xml:",omitempty"`
	SOAPBodyContentType string `xml:"-"`
}

type SoapBody struct {
	Text   string `xml:",chardata"`
	Inform struct {
		Text     string `xml:",chardata"`
		DeviceId struct {
			Text         string `xml:",chardata"`
			Manufacturer string `xml:"Manufacturer"`
			OUI          string `xml:"OUI"`
			ProductClass string `xml:"ProductClass"`
			SerialNumber string `xml:"SerialNumber"`
		} `xml:"DeviceId"`
		Event struct {
			Text        string `xml:",chardata"`
			Type        string `xml:"type,attr"`
			ArrayType   string `xml:"arrayType,attr"`
			EventStruct struct {
				Text       string `xml:",chardata"`
				EventCode  string `xml:"EventCode"`
				CommandKey string `xml:"CommandKey"`
			} `xml:"EventStruct"`
		} `xml:"Event"`
		MaxEnvelopes  string `xml:"MaxEnvelopes"`
		CurrentTime   string `xml:"CurrentTime"`
		RetryCount    string `xml:"RetryCount"`
		ParameterList struct {
			Text                 string `xml:",chardata"`
			Type                 string `xml:"type,attr"`
			ArrayType            string `xml:"arrayType,attr"`
			ParameterValueStruct []struct {
				Text  string `xml:",chardata"`
				Name  string `xml:"Name"`
				Value struct {
					Text string `xml:",chardata"`
					Type string `xml:"type,attr"`
				} `xml:"Value"`
			} `xml:"ParameterValueStruct"`
		} `xml:"ParameterList"`
	} `xml:"Inform"`
}

// Fault type
type Fault struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Fault"`

	Code   string `xml:"faultcode,omitempty"`
	String string `xml:"faultstring,omitempty"`
	Actor  string `xml:"faultactor,omitempty"`
	Detail string `xml:"detail,omitempty"`
}

type rootElem struct {
	root map[string]any
}

func (x *rootElem) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	x.root = map[string]any{"_": start.Name.Local}
	path := []map[string]any{x.root}
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		switch elem := token.(type) {
		case xml.StartElement:
			newMap := map[string]any{"_": elem.Name.Local}
			path[len(path)-1][elem.Name.Local] = newMap
			path = append(path, newMap)
		case xml.EndElement:
			path = path[:len(path)-1]
		case xml.CharData:
			val := strings.TrimSpace(string(elem))
			if val == "" {
				break
			}
			curName := path[len(path)-1]["_"].(string)
			path[len(path)-2][curName] = typeConvert(val)
		}
	}
}

func typeConvert(s string) any {
	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return f
	}
	return s
}

type SoapSessionInfo struct {
	ContentType string
	SoapVersion []byte
	Env         EnvInfo
	EventCodes  map[int]struct{}
}

func NewSoapSessionInfo() *SoapSessionInfo {
	return &SoapSessionInfo{
		ContentType: SoapContentType11,
		SoapVersion: BNamespaceSoap11,
		Env:         EnvInfo{},
	}
}
