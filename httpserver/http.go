package httpserver

import (
	"encoding/xml"
	"fmt"
	"net/http"

	logger "github.com/ecpartan/soap-server-tr069/log"
)

// XMLMarshaller lets you inject your favourite custom xml implementation
type XMLMarshaller interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(xml []byte, v interface{}) error
}

type DefaultMarshaller struct{}

func newDefaultMarshaller() XMLMarshaller {
	return &DefaultMarshaller{}
}

func (dm *DefaultMarshaller) Marshal(v interface{}) ([]byte, error) {
	return xml.MarshalIndent(v, "", "	")
}

func (dm *DefaultMarshaller) Unmarshal(xmlBytes []byte, v interface{}) error {
	return xml.Unmarshal(xmlBytes, v)
}

type ResponseWriter struct {
	w             http.ResponseWriter
	outputStarted bool
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w: w}
}

func (w *ResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	w.outputStarted = true
	logger.LogDebug("writing response: ", string(b))

	return w.w.Write(b)
}

func (w *ResponseWriter) WriteHeader(code int) {
	w.w.WriteHeader(code)
}

// WriteHeader first set the content-type header and then writes the header code.
func WriteHeader(w http.ResponseWriter, ContentType string, code int) {
	setContentType(w, ContentType)
	w.WriteHeader(code)
}

func setContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}

func addSOAPHeader(w http.ResponseWriter, contentLength int, contentType string) {
	setContentType(w, contentType)
	w.Header().Set("Content-Length", fmt.Sprint(contentLength))
}

func TransmitXMLReq(request interface{}, w http.ResponseWriter, contentType string) {
	xmlBytes, err := newDefaultMarshaller().Marshal(request)
	// Adjust namespaces for SOAP 1.2

	if err != nil {
		HandleError(fmt.Errorf("could not marshal response:: %s", err), w)
	}
	addSOAPHeader(w, len(xmlBytes), contentType)
	w.Write(xmlBytes)

}
func HandleError(err error, w http.ResponseWriter) {
	// has to write a soap fault
	logger.LogDebug("handling error:", err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Error " + err.Error()))

	return
}
