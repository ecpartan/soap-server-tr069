package httpserver

import (
	"encoding/xml"
	"fmt"
	"net/http"

	logger "github.com/ecpartan/soap-server-tr069/log"
)

// XMLMarshaller lets you inject your favourite custom xml implementation
type XMLMarshaller interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(xml []byte, v any) error
}

type DefaultMarshaller struct{}

func newDefaultMarshaller() XMLMarshaller {
	return &DefaultMarshaller{}
}

func (dm *DefaultMarshaller) Marshal(v interface{}) ([]byte, error) {
	return xml.MarshalIndent(v, "", "	")
}

func (dm *DefaultMarshaller) Unmarshal(xmlBytes []byte, v any) error {
	return xml.Unmarshal(xmlBytes, v)
}

type ResponseWriter struct {
	W             http.ResponseWriter
	OutputStarted bool
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{W: w, OutputStarted: false}
}

func (w *ResponseWriter) Header() http.Header {
	return w.W.Header()
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	w.OutputStarted = true
	logger.LogDebug("writing response: ", string(b))

	return w.W.Write(b)
}

func (w *ResponseWriter) WriteHeader(code int) {
	w.W.WriteHeader(code)
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

func TransmitXMLReq(request any, w http.ResponseWriter, contentType string) {
	logger.LogDebug("TransmitXMLReq", request, contentType)
	xmlBytes, err := newDefaultMarshaller().Marshal(request)
	// Adjust namespaces for SOAP 1.2

	if err != nil {
		HandleError(fmt.Errorf("could not marshal response:: %s", err), w)
	}
	addSOAPHeader(w, len(xmlBytes), contentType)
	code, err := w.Write(xmlBytes)
	logger.LogDebug("Error writing response: ", err, len(xmlBytes), code)

}
func HandleError(err error, w http.ResponseWriter) {
	// has to write a soap fault
	logger.LogDebug("handling error:", err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Error " + err.Error()))

	return
}
