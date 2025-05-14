package xml

import "encoding/xml"

// XMLMarshaller lets you inject your favourite custom xml implementation
type XMLMarshaller interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(xml []byte, v interface{}) error
}

type DefaultMarshaller struct{}

func (dm DefaultMarshaller) Marshal(v interface{}) ([]byte, error) {
	return xml.MarshalIndent(v, "", "	")
}

func (dm DefaultMarshaller) Unmarshal(xmlBytes []byte, v interface{}) error {
	return xml.Unmarshal(xmlBytes, v)
}
